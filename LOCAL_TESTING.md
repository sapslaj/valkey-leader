# Local Testing Guide

This guide shows how to test the Valkey Leader Helm chart locally using KinD (Kubernetes in Docker).

## Prerequisites

Make sure you have the following tools installed:

- [Docker](https://docs.docker.com/get-docker/)
- [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)

## Quick Start

To run the complete test suite with a single command:

```bash
make test-quick
```

This will:
1. Create a KinD cluster
2. Build the Docker image
3. Load the image into the cluster
4. Install the Helm chart
5. Run basic Valkey operation tests

## Step-by-Step Testing

### 1. Create KinD Cluster

```bash
make kind-create
```

### 2. Build and Install

```bash
make helm-test-install
```

This builds the Docker image, loads it into KinD, and installs the Helm chart.

### 3. Run Tests

#### Basic Operations Test
```bash
make test-operations
```

This tests:
- Write operations via the read-write service
- Read operations via all services (read, read-only, read-write)
- Service connectivity
- Replication status

#### Leader Failover Test
```bash
make test-failover
```

This tests:
- Pod deletion and recreation
- Leader election after failover
- Data persistence during failover
- Service availability after failover

#### Complete Test Suite
```bash
make test-all
```

Runs both operation tests and failover tests.

## Monitoring and Debugging

### Check Status
```bash
make kind-status
```

Shows:
- Cluster information
- Pod status
- Service status
- StatefulSet status

### View Logs
```bash
make kind-logs
```

Shows logs from all Valkey Leader containers and Valkey containers.

### Manual Debugging

Connect to the KinD cluster:
```bash
kubectl config use-context kind-valkey-test
```

Get pod information:
```bash
kubectl get pods -l app.kubernetes.io/name=valkey-leader
```

Connect to a specific pod:
```bash
kubectl exec -it valkey-test-0 -c valkey -- redis-cli
```

Test services manually:
```bash
kubectl run redis-cli --rm -i --restart=Never --image=redis:7-alpine -- redis-cli -h valkey-test-rw -p 6379 info
```

## Cleanup

### Clean Up Resources
```bash
make kind-clean
```

Removes the Helm release and related resources but keeps the cluster.

### Delete Cluster
```bash
make kind-delete
```

Completely removes the KinD cluster.

### Reset Everything
```bash
make kind-reset
```

Deletes the cluster and creates a fresh one.

## Available Commands

Run `make help-local-testing` to see all available commands:

```bash
make help-local-testing
```

## Troubleshooting

### KinD Cluster Won't Start
- Make sure Docker is running
- Check if port 6443 is available
- Try deleting existing cluster: `make kind-delete`

### Pods Stuck in Pending
- Check node resources: `kubectl describe nodes`
- Check pod events: `kubectl describe pod <pod-name>`

### Services Not Accessible
- Verify services exist: `kubectl get services`
- Check service endpoints: `kubectl get endpoints`
- Verify pod labels match service selectors

### Image Pull Errors
- Make sure the image was built: `docker images | grep valkey-leader`
- Verify image was loaded into KinD: `docker exec -it valkey-test-control-plane crictl images | grep valkey-leader`

### Tests Failing
- Check pod logs: `make kind-logs`
- Verify cluster status: `make kind-status`
- Try running tests individually to isolate issues

## Custom Configuration

You can modify the test setup by editing:

- `scripts/kind-setup.sh` - KinD cluster configuration
- `scripts/test-valkey-operations.sh` - Basic operation tests
- `scripts/test-failover.sh` - Failover tests
- `Makefile` - Test targets and parameters

## Integration with CI

The local testing setup mirrors the GitHub Actions CI pipeline, so successful local tests should indicate that CI will pass as well.