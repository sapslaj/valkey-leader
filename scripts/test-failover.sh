#!/bin/bash

set -euo pipefail

RELEASE_NAME="${1:-valkey-test}"
NAMESPACE="${2:-default}"

echo "Testing leader failover for release: $RELEASE_NAME in namespace: $NAMESPACE"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed."
    exit 1
fi

# Function to run redis-cli commands
run_redis_cli() {
    local service="$1"
    local command="$2"

    kubectl run redis-cli-test --rm -i --restart=Never --image=redis:7-alpine \
        --namespace="$NAMESPACE" -- \
        redis-cli -h "${RELEASE_NAME}-valkey-leader-${service}" -p 6379 ${command}
}

# Function to get the first pod
get_first_pod() {
    kubectl get pods -l app.kubernetes.io/name=valkey-leader --namespace="$NAMESPACE" \
        -o jsonpath='{.items[0].metadata.name}'
}

# Main test function
main() {
    echo "Starting failover test..."

    # Wait for pods to be ready initially
    echo "Waiting for pods to be ready..."
    kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=valkey-leader \
        --namespace="$NAMESPACE" --timeout=300s

    # Give services a moment to stabilize
    echo "Waiting for services to stabilize..."
    sleep 10

    # Set initial test data
    echo "Setting initial test data..."
    run_redis_cli "rw" "set failover-test-key initial-value"
    run_redis_cli "rw" "set counter 1"

    # Get current pod to delete
    CURRENT_POD=$(get_first_pod)
    echo "Current pod to delete: $CURRENT_POD"

    # Verify data is accessible before failover
    echo "Verifying data before failover..."
    INITIAL_VALUE=$(run_redis_cli "rw" "get failover-test-key" 2>/dev/null | head -1)
    if [ "$INITIAL_VALUE" != "initial-value" ]; then
        echo "Error: Initial data verification failed"
        exit 1
    fi

    # Delete the pod to trigger failover
    echo "Deleting pod to trigger failover: $CURRENT_POD"
    kubectl delete pod "$CURRENT_POD" --namespace="$NAMESPACE"

    # Wait for StatefulSet to recreate the pod
    echo "Waiting for pods to be ready after failover..."
    kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=valkey-leader \
        --namespace="$NAMESPACE" --timeout=300s

    # Give some time for leader election to complete
    echo "Waiting for leader election to complete..."
    sleep 30

    # Test that service still works after failover
    echo "Testing service after failover..."
    run_redis_cli "rw" "set post-failover-key success"

    # Verify old data is still accessible
    PRESERVED_VALUE=$(run_redis_cli "rw" "get failover-test-key" 2>/dev/null | head -1)
    if [ "$PRESERVED_VALUE" != "initial-value" ]; then
        echo "Warning: Data may not have been preserved during failover"
        echo "Expected: 'initial-value', Got: '$PRESERVED_VALUE'"
    else
        echo "✅ Data preserved during failover"
    fi

    # Verify new data was written successfully
    NEW_VALUE=$(run_redis_cli "rw" "get post-failover-key" 2>/dev/null | head -1)
    if [ "$NEW_VALUE" != "success" ]; then
        echo "Error: Post-failover write failed"
        echo "Expected: 'success', Got: '$NEW_VALUE'"
        exit 1
    fi

    echo "✅ Post-failover write successful"

    # Test increment operation
    run_redis_cli "rw" "incr counter"
    COUNTER_VALUE=$(run_redis_cli "rw" "get counter" 2>/dev/null | head -1)
    echo "Counter value after failover: $COUNTER_VALUE"

    # Check final pod status
    echo "Final pod status:"
    kubectl get pods -l app.kubernetes.io/name=valkey-leader --namespace="$NAMESPACE"

    echo "🎉 Failover test completed successfully!"
}

# Run main function
main "$@"