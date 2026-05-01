.PHONY: help docs check fmt test lint kind-up kind-down
SHELL := pwsh.exe
.SHELLFLAGS := -NoProfile -Command

help: ## Show available commands.
	@Select-String -Path $(MAKEFILE_LIST) -Pattern '^[a-zA-Z0-9_-]+:.*?## ' | ForEach-Object { $$parts = $$_.Line -split ':.*?## ', 2; '{0,-18} {1}' -f $$parts[0], $$parts[1] }

docs: ## Validate that required Phase 0 documentation files exist.
	@$$required = @('README.md','ROADMAP.md','docs/tool-versions.md','docs/ci-plan.md','docs/adr/0001-use-kubernetes-operator.md','docs/adr/0002-use-kafka-as-the-event-log.md','docs/adr/0003-use-openlineage-for-lineage-events.md','docs/adr/0004-keep-lineage-async.md'); $$missing = $$required | Where-Object { -not (Test-Path $$_) }; if ($$missing) { Write-Error ('Missing required docs: ' + ($$missing -join ', ')); exit 1 }

check: docs ## Run all checks currently available in the bootstrap phase.
	@Write-Host "Phase 0 checks passed."

fmt: ## Format project code once Go modules exist.
	@Write-Host "No code to format yet."

test: ## Run tests once implementation packages exist.
	@Write-Host "No tests to run yet."

lint: ## Run linters once implementation packages exist.
	@Write-Host "No linters configured yet."

kind-up: ## Create the local Kubernetes cluster once Phase 1 scripts exist.
	@Write-Host "Local cluster automation starts in Phase 1."

kind-down: ## Delete the local Kubernetes cluster once Phase 1 scripts exist.
	@Write-Host "Local cluster automation starts in Phase 1."
