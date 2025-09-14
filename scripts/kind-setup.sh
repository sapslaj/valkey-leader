#!/bin/bash

set -euo pipefail

CLUSTER_NAME="${1:-valkey-test}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Setting up KinD cluster: $CLUSTER_NAME"

# Check if KinD is installed
if ! command -v kind &> /dev/null; then
    echo "Error: kind is not installed. Please install it first:"
    echo "https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed. Please install it first:"
    echo "https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

# Check if cluster already exists
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "Cluster $CLUSTER_NAME already exists. Deleting it first..."
    kind delete cluster --name "$CLUSTER_NAME"
fi

# Create KinD cluster with configuration
echo "Creating KinD cluster..."
cat <<EOF | kind create cluster --name "$CLUSTER_NAME" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
EOF

# Wait for cluster to be ready
echo "Waiting for cluster to be ready..."
kubectl cluster-info --context "kind-${CLUSTER_NAME}"
kubectl wait --for=condition=Ready nodes --all --timeout=300s

echo "KinD cluster '$CLUSTER_NAME' is ready!"
echo "To use this cluster, run: kubectl config use-context kind-${CLUSTER_NAME}"