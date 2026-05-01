# Phase 2 Security And Admission

Phase 2 is a local, single-tenant operator install in the `kinetix` namespace.
The goal is to make the security boundary explicit while keeping the demo
cluster runnable without cert-manager, external secret stores, or image policy
controllers.

## Implemented Baseline

- `operator/config/default/namespace.yaml` labels the namespace with Pod
  Security `baseline` enforcement and `restricted` audit/warn modes.
- `operator/config/manager/manager.yaml` runs the operator container as
  non-root with `RuntimeDefault` seccomp, no privilege escalation, dropped
  capabilities, and a read-only root filesystem.
- Worker Deployments reconciled by the controller receive the same pod and
  container security posture.
- `operator/config/rbac/pipeline_editor_role.yaml` grants users normal
  namespace-scoped access to `pipelines` but not `pipelines/status` or
  `pipelines/finalizers`.
- `operator/config/network/worker_network_policy.yaml` allows worker metrics
  ingress from the `kinetix` namespace and egress to DNS, Kafka, Redis, Dapr,
  and metrics endpoints used by the local stack.
- `operator/config/network/operator_network_policy.yaml` limits operator
  ingress for health and metrics. Operator egress is intentionally unrestricted
  in Phase 2 because Kubernetes API server addressing differs by cluster.

## Admission Validation

The CRD schema rejects malformed `Pipeline` objects for the stable Phase 2 API
shape. The API package also exposes the same validation logic for controller
status handling and for a controller-runtime validating webhook.

The webhook is registered only when the operator starts with:

```powershell
--enable-validation-webhook
```

Do not enable that flag until webhook TLS serving certificates are installed for
the manager pod. Without certificates, the local Phase 2 operator should rely on
CRD structural validation and reconciler validation.

## Remaining Phase 4 Security Work

- Move the namespace to Pod Security `restricted` enforcement after all injected
  sidecars and local dependency pods satisfy it.
- Add default-deny NetworkPolicies once required Kubernetes API, Strimzi, Dapr,
  Prometheus, and Grafana paths are fully enumerated.
- Add external secret integration for Kafka and future registry credentials.
- Enforce signed operator, worker, and transform images.
- Add SBOM generation and vulnerability scanning to release workflows.
