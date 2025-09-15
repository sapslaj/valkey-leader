#!/bin/bash

set -euo pipefail

CLUSTER_NAME="${1:-valkey-test}"
NAMESPACE="${2:-default}"

echo "Testing leader failover for release: $CLUSTER_NAME in namespace: $NAMESPACE"

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

# Function to get the first pod
get_primary_pod() {
  with_backoff kubectl get pods -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME,valkey.sapslaj.cloud/instance-role=primary" --namespace="$NAMESPACE" \
    -o jsonpath='{.items[0].metadata.name}'
}

# Main test function
main() {
  echo "Starting failover test..."

  # Wait for pods to be ready initially
  echo "Waiting for pods to be ready..."
  kubectl wait --for=condition=Ready pod -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME" \
    --namespace="$NAMESPACE" --timeout=300s

  # Give some time for leader election to complete
  echo "Waiting for leader election to complete..."
  with_backoff kubectl wait --for=condition=Ready pod -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME,valkey.sapslaj.cloud/instance-role=primary" \
    --namespace="$NAMESPACE" --timeout=60s

  # Set initial test data
  echo "Setting initial test data..."
  with_backoff run_redis_cli "rw" "set failover-test-key initial-value"
  with_backoff run_redis_cli "rw" "set counter 1"

  # Get current pod to delete
  CURRENT_POD=$(get_primary_pod)
  echo "Current pod to delete: $CURRENT_POD"

  # Verify data is accessible before failover
  echo "Verifying data before failover..."
  INITIAL_VALUE=$(with_backoff run_redis_cli "rw" "get failover-test-key" | head -1)
  if [ "$INITIAL_VALUE" != "initial-value" ]; then
    echo "Error: Initial data verification failed"
    exit 1
  fi

  # Delete the pod to trigger failover
  echo "Deleting pod to trigger failover: $CURRENT_POD"
  with_backoff kubectl delete pod "$CURRENT_POD" --namespace="$NAMESPACE"

  # Wait for StatefulSet to recreate the pod
  echo "Waiting for pods to be ready after failover..."
  kubectl wait --for=condition=Ready pod -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME" \
    --namespace="$NAMESPACE" --timeout=300s

  # Give some time for leader election to complete
  echo "Waiting for leader election to complete..."
  with_backoff kubectl wait --for=condition=Ready pod -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME,valkey.sapslaj.cloud/instance-role=primary" \
    --namespace="$NAMESPACE" --timeout=60s

  # Test that service still works after failover
  echo "Testing service after failover..."
  with_backoff run_redis_cli "rw" "set post-failover-key success"

  # Verify old data is still accessible
  PRESERVED_VALUE=$(with_backoff run_redis_cli "rw" "get failover-test-key" 2>/dev/null | head -1)
  if [ "$PRESERVED_VALUE" != "initial-value" ]; then
    echo "Warning: Data may not have been preserved during failover"
    echo "Expected: 'initial-value', Got: '$PRESERVED_VALUE'"
  else
    echo "✅ Data preserved during failover"
  fi

  # Verify new data was written successfully
  NEW_VALUE=$(with_backoff run_redis_cli "rw" "get post-failover-key" 2>/dev/null | head -1)
  if [ "$NEW_VALUE" != "success" ]; then
    echo "Error: Post-failover write failed"
    echo "Expected: 'success', Got: '$NEW_VALUE'"
    exit 1
  fi

  echo "✅ Post-failover write successful"

  # Test increment operation
  with_backoff run_redis_cli "rw" "incr counter"
  COUNTER_VALUE=$(with_backoff run_redis_cli "rw" "get counter" 2>/dev/null | head -1)
  echo "Counter value after failover: $COUNTER_VALUE"

  # Check final pod status
  echo "Final pod status:"
  kubectl get pods -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME" --namespace="$NAMESPACE"

  echo "🎉 Failover test completed successfully!"
}

# Run main function
main "$@"
