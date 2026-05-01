# Phase 2 Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the Phase 2 roadmap gaps for validation, RBAC, tenancy, security posture, schema strategy, tests, and documentation.

**Architecture:** Keep the existing controller-runtime operator and add narrowly scoped admission validation, security defaults, namespace-scoped user RBAC, and local NetworkPolicies. Capture roadmap decisions in ADRs and make the docs gate require the new Phase 2 artifacts.

**Tech Stack:** Go, controller-runtime, Kubernetes RBAC, NetworkPolicy, Pod Security Standards, Kustomize, Markdown ADRs.

---

### Task 1: Shared Pipeline Validation

**Files:**
- Create: `operator/api/v1alpha1/pipeline_validation.go`
- Create: `operator/api/v1alpha1/pipeline_validation_test.go`
- Modify: `operator/internal/controller/pipeline_controller.go`
- Modify: `operator/main.go`

- [x] **Step 1: Move validation to the API package**

Create `ValidatePipelineSpec` so CRD-compatible validation, reconciler status handling, and the optional webhook use the same rules.

- [x] **Step 2: Add validating webhook wiring**

Register `PipelineValidator` behind `--enable-validation-webhook` so local installs do not require TLS serving certs by default.

- [x] **Step 3: Test valid and invalid specs**

Run `go test ./...` in `operator` and verify API validation tests pass.

### Task 2: Phase 2 Security Baseline

**Files:**
- Modify: `operator/internal/controller/pipeline_controller.go`
- Modify: `operator/internal/controller/pipeline_controller_test.go`
- Modify: `operator/config/default/namespace.yaml`
- Modify: `operator/config/manager/manager.yaml`
- Create: `operator/config/rbac/pipeline_editor_role.yaml`
- Create: `operator/config/network/operator_network_policy.yaml`
- Create: `operator/config/network/worker_network_policy.yaml`
- Modify: `operator/config/kustomization.yaml`

- [x] **Step 1: Apply pod security defaults**

Set non-root execution, explicit runtime UIDs, `RuntimeDefault` seccomp, no privilege escalation, dropped capabilities, and read-only root filesystems for operator and worker pods.

- [x] **Step 2: Add user RBAC and NetworkPolicies**

Add a namespace-scoped `kinetix-pipeline-editor` role without status/finalizer permissions and add Phase 2 operator/worker NetworkPolicies.

- [x] **Step 3: Verify manifest rendering**

Run `kubectl kustomize operator/config` and verify the new resources render.

### Task 3: Roadmap Documentation

**Files:**
- Create: `docs/adr/0008-use-single-tenant-namespace-for-phase-2.md`
- Create: `docs/adr/0009-phase-2-security-baseline.md`
- Create: `docs/adr/0010-phase-2-schema-registry-strategy.md`
- Create: `docs/phase-2-security-and-admission.md`
- Modify: `docs/phase-2-operator-plan.md`
- Modify: `README.md`
- Modify: `Makefile`

- [x] **Step 1: Document tenancy, security, and schema decisions**

Add ADRs for single-tenant Phase 2 operation, the security baseline, and the schema registry strategy.

- [x] **Step 2: Add an operator security guide**

Document implemented RBAC, pod security, NetworkPolicies, admission behavior, and remaining Phase 4 work.

- [x] **Step 3: Extend docs checks**

Make `make docs` require the Phase 2 docs, ADRs, RBAC, and NetworkPolicy artifacts.

### Task 4: Verification

**Files:**
- Verify: repository root

- [x] **Step 1: Run operator tests**

Run `go test ./...` from `operator`.

- [x] **Step 2: Run full checks**

Run `make check` from the repository root and confirm worker tests, operator tests, and docs checks pass.
