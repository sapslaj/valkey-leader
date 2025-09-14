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

# Local testing with KinD
.PHONY: kind-create
kind-create:
	@echo "Creating KinD cluster for testing..."
	./scripts/kind-setup.sh valkey-test

.PHONY: kind-delete
kind-delete:
	@echo "Deleting KinD cluster..."
	./scripts/kind-cleanup.sh valkey-test

.PHONY: kind-test-setup
kind-test-setup: kind-create build-docker
	@echo "Setting up test environment in KinD cluster..."
	kind load docker-image localhost/valkey-leader:dev --name valkey-test
	kubectl config use-context kind-valkey-test

.PHONY: helm-test-install
helm-test-install: kind-test-setup
	@echo "Installing Helm chart for testing..."
	helm upgrade --install valkey-test ./helm/valkey-leader \
		--set valkeyLeader.image.repository=localhost/valkey-leader \
		--set valkeyLeader.image.tag=dev \
		--set valkeyLeader.image.pullPolicy=Never \
		--wait --timeout=300s
	@echo "Waiting for services to stabilize..."
	@sleep 10

.PHONY: test-operations
test-operations:
	@echo "Testing Valkey operations..."
	./scripts/test-valkey-operations.sh valkey-test default

.PHONY: test-simple
test-simple:
	@echo "Running simple Valkey test..."
	./scripts/simple-test.sh valkey-test default

.PHONY: test-failover
test-failover:
	@echo "Testing leader failover..."
	./scripts/test-failover.sh valkey-test default

.PHONY: test-all
test-all: helm-test-install test-operations test-failover
	@echo "All tests completed successfully!"

.PHONY: test-quick
test-quick: helm-test-install test-simple
	@echo "Quick test completed successfully!"

.PHONY: kind-logs
kind-logs:
	@echo "Getting logs from all valkey-leader pods..."
	@for pod in $$(kubectl get pods -l app.kubernetes.io/name=valkey-leader -o jsonpath='{.items[*].metadata.name}'); do \
		echo "=== Logs for $$pod (valkey-leader container) ==="; \
		kubectl logs $$pod -c valkey-leader --tail=50 || true; \
		echo "=== Logs for $$pod (valkey container) ==="; \
		kubectl logs $$pod -c valkey --tail=50 || true; \
		echo ""; \
	done

.PHONY: kind-status
kind-status:
	@echo "=== Cluster Info ==="
	kubectl cluster-info
	@echo ""
	@echo "=== Pods ==="
	kubectl get pods -l app.kubernetes.io/name=valkey-leader
	@echo ""
	@echo "=== Services ==="
	kubectl get services -l app.kubernetes.io/name=valkey-leader
	@echo ""
	@echo "=== StatefulSet ==="
	kubectl get statefulset -l app.kubernetes.io/name=valkey-leader

.PHONY: kind-clean
kind-clean:
	@echo "Cleaning up test resources..."
	helm uninstall valkey-test || true
	kubectl delete pods -l app.kubernetes.io/name=valkey-leader || true

.PHONY: kind-reset
kind-reset: kind-clean kind-delete kind-create
	@echo "KinD cluster reset complete"

.PHONY: help-local-testing
help-local-testing:
	@echo "Local Testing Commands:"
	@echo "  /usr/bin/make kind-create         - Create KinD cluster for testing"
	@echo "  /usr/bin/make kind-delete         - Delete KinD cluster"
	@echo "  /usr/bin/make kind-test-setup     - Create cluster and load Docker image"
	@echo "  /usr/bin/make helm-test-install   - Install Helm chart in test cluster"
	@echo "  /usr/bin/make test-simple         - Run simple Valkey operations test (recommended)"
	@echo "  /usr/bin/make test-operations     - Run detailed Valkey operations tests"
	@echo "  /usr/bin/make test-failover       - Run leader failover tests"
	@echo "  /usr/bin/make test-all           - Run complete test suite (install + all tests)"
	@echo "  /usr/bin/make test-quick         - Run quick test suite (install + simple test)"
	@echo "  /usr/bin/make kind-logs          - Show logs from all pods"
	@echo "  /usr/bin/make kind-status        - Show cluster status"
	@echo "  /usr/bin/make kind-clean         - Clean up test resources"
	@echo "  /usr/bin/make kind-reset         - Reset entire test environment"
	@echo ""
	@echo "Quick Start:"
	@echo "  /usr/bin/make test-quick         - One command to set up and test everything"
	@echo ""
	@echo "Note: Use '/usr/bin/make' if 'make' conflicts with shell functions"
