# valkey-leader

> [!NOTE]
> Disclaimer: Major parts of this project were written with AI/LLMs, notably
> Anthropic Claude.

valkey-leader is a little sidecar container for Valkey that uses
Kubernetes-native leader election to determine which instance is the primary
and which are the replicas. No Raft, no Sentinel, no Cluster mode. Just simple
resource locks built into Kubernetes.

## Usage

### Helm chart

The easiest way is using the pre-packaged Helm chart that includes:

- A Valkey StatefulSet
- valkey-leader
- [redis_exporter](https://github.com/oliver006/redis_exporter)

#### Install from GHCR

```bash
# Install latest version
helm install my-valkey oci://ghcr.io/sapslaj/valkey-leader-chart/valkey-leader

# Install in a specific namespace
helm install my-valkey oci://ghcr.io/sapslaj/valkey-leader-chart/valkey-leader \
  --namespace valkey --create-namespace

# Install specific version
helm install my-valkey oci://ghcr.io/sapslaj/valkey-leader-chart/valkey-leader --version 0.1.0
```

#### Local Development

```bash
# Clone the repository
git clone https://github.com/sapslaj/valkey-leader.git
cd valkey-leader

# Install with local chart (production images)
helm install my-valkey ./helm/valkey-leader

# Install with local chart and local development image
make helm-deploy
```

#### Configuration

The Helm chart supports extensive configuration. Key options include:

```yaml
# values.yaml
valkeyLeader:
  image:
    repository: ghcr.io/sapslaj/valkey-leader
    tag: latest

redisExporter:
  enabled: true # Enable Prometheus metrics

monitoring:
  serviceMonitor:
    enabled: true # Create ServiceMonitor for Prometheus Operator

statefulSet:
  replicas: 3 # Number of Valkey instances
```

### Integrate into existing Valkey setup

valkey-leader is meant to be run as a sidecar container to Valkey. It
communicates over `localhost:6379` to control Valkey. Currently it does not
support authentication or TLS (will be added later). For an example of how to
set this up see `./deploy/base/statefulset.yaml`.

Configuration is done via environment variables.

| Environment Variable | Required | Description                                               | Example Value                                 |
| -------------------- | -------- | --------------------------------------------------------- | --------------------------------------------- |
| `CLUSTER_NAME`       | Yes      | Name of the Valkey cluster for leader election            | `my-valkey-cluster`                           |
| `NAMESPACE`          | Yes      | Kubernetes namespace where the pods are running           | `default`                                     |
| `POD_IP`             | Yes      | IP address of the current pod                             | `10.244.0.5`                                  |
| `POD_NAME`           | Yes      | Name of the current pod                                   | `my-valkey-0`                                 |
| `SERVICE_NAME`       | Yes      | Name of the headless service for pod discovery            | `my-valkey-headless`                          |
| `LEADER_LEASE_NAME`  | No       | Name of the Kubernetes lease resource for leader election | `my-valkey-leader` (defaults to cluster name) |

In order for valkey-leader's leader election to work correctly, it needs the
RBAC permissions outlined in `./deploy/base/role.yaml`. If the Valkey workload
does not have a ServiceAccount, one will need to be created and a new
RoleBinding added for the Role.

valkey-leader adds the labels `valkey.sapslaj.cloud/cluster` and
`valkey.sapslaj.cloud/instance-role` to the Pods. These can be used in label
selectors to find primaries, replicas, or both.
