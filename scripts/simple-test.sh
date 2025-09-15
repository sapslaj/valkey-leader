#!/bin/bash

set -euo pipefail

CLUSTER_NAME="${1:-valkey-test}"
NAMESPACE="${2:-default}"

echo "Running simple Valkey operations test for release: $CLUSTER_NAME in namespace: $NAMESPACE"

# Wait for pods to be ready
echo "Waiting for pods to be ready..."
kubectl wait --for=condition=Ready pod -l "valkey.sapslaj.cloud/cluster=$CLUSTER_NAME" \
  --namespace="$NAMESPACE" --timeout=300s

# Run comprehensive test in a single pod to avoid script complexity
echo "Running comprehensive test..."
kubectl run simple-test --rm -i --restart=Never --image=redis:7-alpine \
  --namespace="$NAMESPACE" -- sh -c "
set -e

echo 'Testing write operations...'
redis-cli -h ${CLUSTER_NAME}-rw -p 6379 set simple-test-key success
redis-cli -h ${CLUSTER_NAME}-rw -p 6379 set counter 1
redis-cli -h ${CLUSTER_NAME}-rw -p 6379 incr counter

echo 'Testing read operations from read-write service...'
VALUE=\$(redis-cli -h ${CLUSTER_NAME}-rw -p 6379 get simple-test-key)
if [ \"\$VALUE\" != \"success\" ]; then
    echo \"Error: Expected 'success', got '\$VALUE'\"
    exit 1
fi
echo \"✅ Read-write service test passed\"

echo 'Testing read operations from read service...'
VALUE=\$(redis-cli -h ${CLUSTER_NAME}-r -p 6379 get simple-test-key)
if [ \"\$VALUE\" != \"success\" ]; then
    echo \"Error: Expected 'success', got '\$VALUE'\"
    exit 1
fi
echo \"✅ Read service test passed\"

echo 'Testing read operations from read-only service...'
VALUE=\$(redis-cli -h ${CLUSTER_NAME}-ro -p 6379 get simple-test-key)
if [ \"\$VALUE\" != \"success\" ]; then
    echo \"Error: Expected 'success', got '\$VALUE'\"
    exit 1
fi
echo \"✅ Read-only service test passed\"

echo 'Checking counter value...'
COUNTER=\$(redis-cli -h ${CLUSTER_NAME}-r -p 6379 get counter)
if [ \"\$COUNTER\" != \"2\" ]; then
    echo \"Error: Expected counter '2', got '\$COUNTER'\"
    exit 1
fi
echo \"✅ Counter test passed\"

echo 'Testing service connectivity...'
for service in headless r rw ro; do
    if [ \"\$service\" = \"headless\" ]; then
        echo \"Skipping ping test for headless service\"
        continue
    fi
    PING=\$(redis-cli -h ${CLUSTER_NAME}-\$service -p 6379 ping)
    if [ \"\$PING\" != \"PONG\" ]; then
        echo \"Error: Service \$service ping failed, got '\$PING'\"
        exit 1
    fi
    echo \"✅ Service \$service connectivity test passed\"
done

echo '🎉 All tests passed successfully!'
"

echo "Simple test completed successfully!"
