.PHONY: help docs check fmt test operator-fmt operator-test lint kind-up kind-down deps-up docker-build docker-load deploy-demo smoke
SHELL := pwsh.exe
.SHELLFLAGS := -NoProfile -Command

help: ## Show available commands.
	@Select-String -Path $(MAKEFILE_LIST) -Pattern '^[a-zA-Z0-9_-]+:.*?## ' | ForEach-Object { $$parts = $$_.Line -split ':.*?## ', 2; '{0,-18} {1}' -f $$parts[0], $$parts[1] }

docs: ## Validate that required documentation files exist.
	@$$required = @('README.md','ROADMAP.md','docs/tool-versions.md','docs/ci-plan.md','docs/phase-1-local-runtime-plan.md','docs/phase-2-operator-plan.md','docs/phase-2-security-and-admission.md','docs/adr/0001-use-kubernetes-operator.md','docs/adr/0002-use-kafka-as-the-event-log.md','docs/adr/0003-use-openlineage-for-lineage-events.md','docs/adr/0004-keep-lineage-async.md','docs/adr/0005-use-controller-runtime-for-phase-2-operator.md','docs/adr/0006-reconcile-pipeline-to-one-worker-deployment.md','docs/adr/0007-manage-kafka-topics-with-strimzi-kafkatopic.md','docs/adr/0008-use-single-tenant-namespace-for-phase-2.md','docs/adr/0009-phase-2-security-baseline.md','docs/adr/0010-phase-2-schema-registry-strategy.md','operator/config/rbac/pipeline_editor_role.yaml','operator/config/network/operator_network_policy.yaml','operator/config/network/worker_network_policy.yaml'); $$missing = $$required | Where-Object { -not (Test-Path $$_) }; if ($$missing) { Write-Error ('Missing required docs: ' + ($$missing -join ', ')); exit 1 }

check: docs test operator-test ## Run documentation and code checks.
	@Write-Host "Checks passed."

fmt: ## Format Go code.
	@Set-Location workers; gofmt -w ./cmd ./internal

test: ## Run Go unit tests.
	@Set-Location workers; go test ./...

operator-fmt: ## Format operator Go code.
	@Set-Location operator; gofmt -w .

operator-test: ## Run operator Go tests.
	@Set-Location operator; go test ./...

lint: ## Run linters once implementation packages exist.
	@Write-Host "No linters configured yet."

kind-up: ## Create the local kind cluster.
	@./scripts/kind-up.ps1

kind-down: ## Delete the local kind cluster.
	@./scripts/kind-down.ps1

deps-up: ## Install local runtime dependencies into the kind cluster.
	@./scripts/deps-up.ps1

docker-build: ## Build the local demo worker image.
	@./scripts/docker-build.ps1

docker-load: ## Load the local demo image into kind.
	@./scripts/docker-load.ps1

deploy-demo: ## Deploy the hardcoded Phase 1 pipeline.
	@./scripts/deploy-demo.ps1

smoke: ## Verify records flow from source to sink.
	@./scripts/smoke.ps1
