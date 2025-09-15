#!/bin/bash

set -euo pipefail

CLUSTER_NAME="${1:-valkey-test}"

echo 'apiVersion: valkey-leader.sapslaj.cloud/v1alpha1'
echo 'kind: Valkey'
echo 'metadata:'
echo "  name: $CLUSTER_NAME"
echo 'spec:'
echo '  replicas: 2'
echo '  valkeyLeader:'
echo '    imageTag: dev'
