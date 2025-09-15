#!/usr/bin/env -S just --justfile

cluster_name := "valkey-test"

docker:
  go tool github.com/TecharoHQ/yeet/cmd/yeet -fname ./yeetfile.docker.js

go:
  go tool github.com/TecharoHQ/yeet/cmd/yeet -fname ./yeetfile.go.js

yoke:
  go tool github.com/TecharoHQ/yeet/cmd/yeet -fname ./yeetfile.yoke.js

kind-create:
  ./scripts/kind-setup.sh '{{ cluster_name }}'

kind-cleanup:
  ./scripts/kind-cleanup.sh '{{ cluster_name }}'

kind-load-image:
  DOCKER_TAGS=dev go tool github.com/TecharoHQ/yeet/cmd/yeet -fname ./yeetfile.docker.js
  kind load docker-image -n '{{ cluster_name }}' ghcr.io/sapslaj/valkey-leader:dev
  -kubectl get statefulset '{{ cluster_name }}' && kubectl rollout restart 'statefulset/{{ cluster_name }}'

helm-test-install: kind-load-image
  helm upgrade --install '{{ cluster_name }}' ./helm/valkey-leader \
    --set valkeyLeader.image.tag=dev \
    --set valkeyLeader.image.pullPolicy=Never \
    --set 'fullnameOverride={{ cluster_name }}' \
    --wait --timeout=300s
  kubectl wait \
    --for=condition=Ready pod \
    -l valkey.sapslaj.cloud/cluster=valkey-test \
    --timeout=300s

helm-test-uninstall:
  -helm uninstall '{{ cluster_name }}'
  -kubectl get pods -l 'valkey.sapslaj.cloud/cluster={{ cluster_name }}'

yoke-test-install: kind-load-image
  ./scripts/yoke-test-cr.sh | go run ./yoke/v1/valkey/flight | go tool github.com/yokecd/yoke/cmd/yoke apply '{{ cluster_name }}'

yoke-test-uninstall:
  go tool github.com/yokecd/yoke/cmd/yoke delete '{{ cluster_name }}'

kind-simple-test:
  CLUSTER_NAME='{{ cluster_name }}' ./scripts/simple-test.sh

kind-test-failover:
  CLUSTER_NAME='{{ cluster_name }}' ./scripts/test-failover.sh

kind-test-valkey-operations:
  CLUSTER_NAME='{{ cluster_name }}' ./scripts/test-valkey-operations.sh
