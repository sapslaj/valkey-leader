# CI/CD Setup

This project uses GitHub Actions for continuous integration and delivery.

## Workflows

### 1. Continuous Integration (`ci.yml`)
Runs on pull requests and validates:
- Go code compilation and tests
- Helm chart linting and templating
- Docker build (without push)

### 2. Build Binary Image (`build-image.yml`)
Builds and pushes the `valkey-leader` binary as a container image to `ghcr.io/sapslaj/valkey-leader`:

**Triggers:**
- `main` branch push → tagged with short SHA + `latest`
- Git tag push (`v*`) → tagged with the git tag version
- Pull request → build only (no push)

**Features:**
- Multi-architecture builds (linux/amd64, linux/arm64)
- GitHub Container Registry (GHCR) publishing
- Build provenance attestation
- Layer caching for faster builds

### 3. Build Helm Chart (`build-helm-chart.yml`)
Packages and pushes the Helm chart as an OCI artifact to `ghcr.io/sapslaj/valkey-leader-chart`:

**Triggers:**
- `main` branch push with helm chart changes → versioned with short SHA + `latest`
- Git tag push (`v*`) with helm chart changes → versioned with git tag
- Pull request → lint and test only

**Features:**
- Helm chart linting and validation
- Dynamic versioning based on git context
- OCI registry publishing

## Image Tags

### Binary Image (`ghcr.io/sapslaj/valkey-leader`)
- `latest` - Latest main branch build
- `<short-sha>` - Specific commit builds
- `<version>` - Release builds (e.g., `v1.0.0`)

### Helm Chart (`ghcr.io/sapslaj/valkey-leader-chart`)
- `latest` - Latest main branch chart
- `0.1.0-<short-sha>` - Development builds
- `<version>` - Release builds (e.g., `1.0.0`)

## Usage

### Install from Registry
```bash
# Install latest chart with latest image
helm install my-valkey oci://ghcr.io/sapslaj/valkey-leader-chart

# Install specific version
helm install my-valkey oci://ghcr.io/sapslaj/valkey-leader-chart --version 1.0.0

# Use specific binary image version
helm install my-valkey oci://ghcr.io/sapslaj/valkey-leader-chart \
  --set valkeyLeader.image.tag=abc1234
```

### Local Development
```bash
# Build and deploy locally
make helm-deploy

# Deploy using production image with local chart
make helm-deploy-production

# Install directly from registry
make helm-install-from-registry
```

## Security

- All images are built with distroless/minimal base images
- Multi-stage builds to reduce attack surface
- Build provenance attestation for supply chain security
- GHCR integration with GitHub's security scanning
- No secrets or credentials embedded in images
