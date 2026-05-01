# Phase 2 Operator Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the hardcoded Phase 1 worker deployment with a declarative `Pipeline` custom resource reconciled by an operator.

**Architecture:** Add a focused Go module in `/operator` using controller-runtime patterns without requiring Kubebuilder to be installed locally. The controller owns Kubernetes resources for worker configuration and runtime execution, records status conditions, and uses a finalizer for cleanup.

**Tech Stack:** Go, controller-runtime, Kubernetes API machinery, envtest/fake clients, Strimzi `KafkaTopic` manifests, kustomize-compatible YAML.

---

### Task 1: Operator Module And API

**Files:**
- Create: `operator/go.mod`
- Create: `operator/api/v1alpha1/pipeline_types.go`
- Create: `operator/api/v1alpha1/groupversion_info.go`
- Create: `operator/internal/controller/pipeline_controller.go`

- [ ] **Step 1: Create the operator Go module**

Run:

```powershell
Set-Location operator
go mod init github.com/kinetix/kinetix-operator/operator
go get sigs.k8s.io/controller-runtime@latest
go get k8s.io/api@latest k8s.io/apimachinery@latest k8s.io/client-go@latest
```

Expected: `operator/go.mod` and `operator/go.sum` are created.

- [ ] **Step 2: Define the `Pipeline` API**

Create `PipelineSpec` with `source`, `transforms`, and `sink`. Create `PipelineStatus` with `observedGeneration`, `workerDeployment`, `configMap`, and `conditions`.

- [ ] **Step 3: Implement the controller skeleton**

Add reconcile constants, finalizer handling, and `SetupWithManager` watching `Pipeline`, `Deployment`, and `ConfigMap`.

### Task 2: Reconciliation Behavior

**Files:**
- Modify: `operator/internal/controller/pipeline_controller.go`
- Test: `operator/internal/controller/pipeline_controller_test.go`

- [ ] **Step 1: Write tests for owned resources**

Use controller-runtime fake client to verify a valid `Pipeline` creates one ConfigMap and one Deployment with owner references and worker-compatible environment variables.

- [ ] **Step 2: Implement desired ConfigMap**

Populate `KINETIX_BROKERS`, `KINETIX_INPUT_TOPIC`, `KINETIX_OUTPUT_TOPIC`, `KINETIX_GROUP`, `KINETIX_SOURCE_SCHEMA`, and transform metadata.

- [ ] **Step 3: Implement desired Deployment**

Create one worker Deployment per `Pipeline`, use the first transform image, set replicas, add Dapr and Prometheus annotations, mount environment from the ConfigMap, and set owner references.

- [ ] **Step 4: Implement idempotent updates**

On repeated reconciliation, update changed ConfigMap data, Deployment image, replicas, labels, annotations, and pod template without creating duplicate resources.

### Task 3: Status, Finalizer, And Manifests

**Files:**
- Modify: `operator/internal/controller/pipeline_controller.go`
- Create: `operator/config/crd/bases/pipeline.kinetix.io_pipelines.yaml`
- Create: `operator/config/rbac/role.yaml`
- Create: `operator/config/rbac/service_account.yaml`
- Create: `operator/config/rbac/role_binding.yaml`
- Create: `operator/config/manager/manager.yaml`
- Create: `operator/config/kustomization.yaml`
- Create: `examples/pipeline.yaml`

- [ ] **Step 1: Add conditions**

Set `Ready=True` after resource reconciliation succeeds. Set `Ready=False` with a useful reason when reconciliation fails or when the spec is invalid.

- [ ] **Step 2: Add finalizer cleanup**

When deleting, remove owned ConfigMap and Deployment if present, then remove `pipeline.kinetix.io/finalizer`.

- [ ] **Step 3: Add installable manifests**

Provide CRD, RBAC, manager Deployment, and an example `Pipeline` that maps Phase 1 `kinetix-input` to `kinetix-output`.

### Task 4: Commands And Verification

**Files:**
- Modify: `Makefile`
- Modify: `README.md`
- Create: `docs/phase-2-operator-plan.md`

- [ ] **Step 1: Add Make targets**

Add `operator-fmt`, `operator-test`, and include operator tests in `check`.

- [ ] **Step 2: Document Phase 2**

Document the API shape, expected resource names, test command, and example apply command.

- [ ] **Step 3: Verify**

Run:

```powershell
Set-Location operator
go test ./...
Set-Location ..
make check
```

Expected: all Go tests and documentation checks pass.
