#!/bin/bash

set -euo pipefail

CLUSTER_NAME="${1:-valkey-test}"

echo "Cleaning up KinD cluster: $CLUSTER_NAME"

# Check if KinD is installed
if ! command -v kind &>/dev/null; then
  echo "Error: kind is not installed."
  exit 1
fi

# Check if cluster exists
if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
  echo "Cluster $CLUSTER_NAME does not exist. Nothing to clean up."
  exit 0
fi

# Delete cluster
echo "Deleting KinD cluster..."
kind delete cluster --name "$CLUSTER_NAME"

echo "KinD cluster '$CLUSTER_NAME' has been deleted."
