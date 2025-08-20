.PHONY: build-docker
build-docker:
	docker build -t localhost/valkey-leader:dev .

.PHONY: kind-deploy
deploy-kind: build-docker
	kind load docker-image localhost/valkey-leader:dev
	kubectl apply -k ./deploy/dev
	kubectl rollout restart -n valkey statefulset/valkey

.PHONY: helm-deploy
helm-deploy: build-docker
	kind load docker-image localhost/valkey-leader:dev
	helm upgrade --install valkey-test ./helm/valkey-leader -n valkey --create-namespace --set valkeyLeader.image.repository=localhost/valkey-leader --set valkeyLeader.image.tag=dev --set valkeyLeader.image.pullPolicy=Never

.PHONY: helm-deploy-production
helm-deploy-production:
	helm upgrade --install valkey-test ./helm/valkey-leader -n valkey --create-namespace

.PHONY: helm-install-from-registry
helm-install-from-registry:
	helm install valkey-test oci://ghcr.io/sapslaj/valkey-leader-chart -n valkey --create-namespace
