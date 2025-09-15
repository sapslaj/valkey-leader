#!/bin/bash

set -euo pipefail

CLUSTER_NAME="${1:-valkey-test}"
NAMESPACE="${2:-default}"

echo "Testing Valkey operations for release: $CLUSTER_NAME in namespace: $NAMESPACE"

function with_backoff {
  local max_attempts=${ATTEMPTS-5}
  local timeout=${TIMEOUT-1}
  local attempt=0
  local exit_code=0

  set +e
  while [[ $attempt -lt $max_attempts ]]; do
    "$@"
    exit_code=$?

    if [[ $exit_code == 0 ]]; then
      set -e
      break
    fi

    echo "Failure running ($*) [$exit_code]; retrying in $timeout." 1>&2
    sleep "$timeout"
    attempt=$((attempt + 1))
    timeout=$((timeout * 2))
  done

  if [[ $exit_code != 0 ]]; then
    echo "Failure running ($*) [$exit_code]; No more retries left." 1>&2
  fi

  set -e
  return $exit_code
}

# Function to run redis-cli commands
run_redis_cli() {
  local service="$1"
  local command="$2"

  if kubectl get pod --namespace "$NAMESPACE" | grep redis-cli-test; then
    kubectl delete --namespace "$NAMESPACE" pod/redis-cli-test || true
  fi
  kubectl run redis-cli-test --rm -i --restart=Never --image=redis:7-alpine \
    --namespace="$NAMESPACE" -- \
    redis-cli -h "${CLUSTER_NAME}-${service}" -p 6379 ${command}
  if kubectl get pod --namespace "$NAMESPACE" | grep redis-cli-test; then
    kubectl delete --namespace "$NAMESPACE" pod/redis-cli-test || true
  fi
}

# Function to test connectivity and basic operations
test_basic_operations() {
  echo "=== Testing basic operations ==="

  echo "Testing write operations via read-write service..."
  with_backoff run_redis_cli "rw" "set test-key hello-world"
  with_backoff run_redis_cli "rw" "set counter 0"
  with_backoff run_redis_cli "rw" "incr counter"

  echo "Testing read operations via different services..."

  echo "Reading from read-write service:"
  if ! run_redis_cli "rw" "get test-key" 2>/dev/null | head -1 | grep -q "hello-world"; then
    echo "Error: Could not read 'hello-world' from read-write service"
    return 1
  else
    echo "✅ Successfully read from read-write service"
  fi

  echo "Reading from read service:"
  if ! run_redis_cli "r" "get test-key" 2>/dev/null | head -1 | grep -q "hello-world"; then
    echo "Error: Could not read 'hello-world' from read service"
    return 1
  else
    echo "✅ Successfully read from read service"
  fi

  echo "Reading from read-only service:"
  if ! run_redis_cli "ro" "get test-key" 2>/dev/null | head -1 | grep -q "hello-world"; then
    echo "Error: Could not read 'hello-world' from read-only service"
    return 1
  else
    echo "✅ Successfully read from read-only service"
  fi

  echo "✅ Basic operations test passed!"
}

# Function to test replication status
test_replication() {
  echo "=== Testing replication status ==="

  echo "Checking replication info from read-write service:"
  with_backoff run_redis_cli "rw" "info replication"

  echo "✅ Replication status checked!"
}

# Function to test services are accessible
test_services() {
  echo "=== Testing service accessibility ==="

  local services=("headless" "r" "rw" "ro")

  for service in "${services[@]}"; do
    local full_service_name="${CLUSTER_NAME}-${service}"
    echo "Testing service: ${full_service_name}"
    if kubectl get service "${full_service_name}" --namespace="$NAMESPACE" >/dev/null 2>&1; then
      echo "✅ Service ${full_service_name} exists"

      # Skip connectivity test for headless service as it works differently
      if [ "$service" == "headless" ]; then
        echo "✅ Service ${full_service_name} is headless (skipping ping test)"
      else
        # Test basic connectivity for regular services
        local ping_result
        ping_result="$(with_backoff run_redis_cli "${service}" "ping")"
        if echo "$ping_result" | grep -q "PONG"; then
          echo "✅ Service ${full_service_name} is accessible"
        else
          echo "❌ Service ${full_service_name} is not accessible (got: '$ping_result')"
          return 1
        fi
      fi
    else
      echo "❌ Service ${full_service_name} does not exist"
      return 1
    fi
  done
}

# Function to test leader election by checking which pod is primary
test_leader_election() {
  echo "=== Testing leader election ==="

  echo "Getting pod information:"
  kubectl get pods -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME" --namespace="$NAMESPACE"

  echo "Checking which pod is the primary by querying the read-write service:"
  with_backoff run_redis_cli "rw" "info server | grep redis_version"

  echo "✅ Leader election information retrieved!"
}

# Main execution
main() {
  echo "Starting Valkey operations test..."

  # Wait for pods to be ready
  echo "Waiting for pods to be ready..."
  kubectl wait --for=condition=Ready pod -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME" \
    --namespace="$NAMESPACE" --timeout=300s

  # Give some time for leader election to complete
  echo "Waiting for leader election to complete..."
  with_backoff kubectl wait --for=condition=Ready pod -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME,valkey.sapslaj.cloud/instance-role=primary" \
    --namespace="$NAMESPACE" --timeout=60s

  # Run tests
  test_services
  test_basic_operations
  test_replication
  test_leader_election

  echo "🎉 All tests passed successfully!"
}

# Run main function
main "$@"
