# Phase 2 Operator Plan

Phase 2 introduces the `Pipeline` custom resource and an operator that turns a
declarative pipeline into the worker runtime resources used by Phase 1.

## API Shape

```yaml
apiVersion: pipeline.kinetix.io/v1alpha1
kind: Pipeline
metadata:
  name: example
  namespace: kinetix
spec:
  source:
    kind: KafkaSource
    topic: kinetix-input
    schema: user-event-v1
  transforms:
    - name: normalize
      image: kinetix/demo:dev
      replicas: 1
  sink:
    kind: KafkaSink
    topic: kinetix-output
```

The first controller slice supports one worker Deployment per `Pipeline` using
the first transform entry. Additional transforms stay in the API as an additive
path for later multi-stage reconciliation.

## Reconciled Resources

For a `Pipeline` named `example`, the operator reconciles:

- `ConfigMap/example-config` with Phase 1 worker environment variables.
- `Deployment/example-worker` running `/worker` from the transform image.
- `KafkaTopic/example-source` for the source topic.
- `KafkaTopic/example-sink` for the sink topic.

All resources receive controller owner references. The `Pipeline` finalizer also
deletes these children explicitly before the API object is removed.

Worker pods are reconciled with the Phase 2 security baseline: non-root pod
execution, `RuntimeDefault` seccomp, no privilege escalation, dropped Linux
capabilities, and a read-only root filesystem.

## Status

`kubectl get pipelines -n kinetix` shows readiness, worker name, source topic,
and sink topic. The controller writes a `Ready` condition:

- `Ready=True`, `Reason=ResourcesReady` after successful reconciliation.
- `Ready=False`, `Reason=InvalidSpec` when the manifest is invalid.
- `Ready=False`, `Reason=ReconcileFailed` when Kubernetes resource writes fail.

## Commands

Run operator tests:

```powershell
make operator-test
```

Run the optional envtest path when Kubebuilder test assets are installed:

```powershell
$env:KINETIX_RUN_ENVTEST='1'
$env:KUBEBUILDER_ASSETS='C:\path\to\kubebuilder\bin'
make operator-test
```

Install the CRD and operator manifests after building and loading
`kinetix/operator:dev` into the local cluster:

```powershell
kubectl apply -k operator/config
kubectl apply -f examples/pipeline.yaml
kubectl get pipelines -n kinetix
```

## Phase 2 Decisions

- [ADR 0008](adr/0008-use-single-tenant-namespace-for-phase-2.md) chooses a
  single-tenant `kinetix` namespace for Phase 2.
- [ADR 0009](adr/0009-phase-2-security-baseline.md) defines the initial
  security baseline.
- [ADR 0010](adr/0010-phase-2-schema-registry-strategy.md) documents the schema
  registry interpretation before Phase 3.
- [Phase 2 security and admission](phase-2-security-and-admission.md) documents
  RBAC, pod security, NetworkPolicies, and webhook behavior.
