# Kinetix Operator

Kinetix Operator is a Kubernetes-native real-time data pipeline platform. The
project is planned around three planes:

- **Control plane:** a Kubernetes operator that reconciles declarative
  `Pipeline` resources into runtime infrastructure.
- **Data plane:** Kafka-backed workers that process records through source,
  transform, and sink stages.
- **Lineage plane:** asynchronous OpenLineage events stored for debugging,
  auditability, and record-level traceability.

The first runtime milestone is intentionally small: run one hardcoded Kafka
pipeline locally before introducing CRDs or the operator reconciler.

## Repository Layout

```text
/operator        Kubernetes operator, API types, controllers, and webhooks
/workers         Worker runtime package, transform examples, and producers
/lineage         Lineage ingestion service and query API
/ui              Pipeline and lineage UI
/charts          Helm charts for local and installable deployments
/config          Kubebuilder manifests and kustomize overlays
/deploy          Local cluster manifests and environment profiles
/docs            Architecture notes, ADRs, diagrams, and operating guides
/scripts         Repeatable local setup, demo, test, and cleanup scripts
/test            End-to-end fixtures and kuttl tests
```

## Local Prerequisites

Phase 0 does not require a running cluster. The expected local toolchain for
the next phases is documented in [docs/tool-versions.md](docs/tool-versions.md).

At a minimum, contributors should expect to install:

- Go
- Docker
- kind or k3d
- kubectl
- Helm
- Kubebuilder
- Dapr CLI
- k9s and stern
- kafkactl or kaf

## Common Commands

```sh
make help
make docs
make check
```

Phase 1 local runtime commands:

```sh
make kind-up
make deps-up
make docker-build
make docker-load
make deploy-demo
make smoke
make kind-down
```

`make deploy-demo` exposes Prometheus at <http://localhost:30000> and Grafana
at <http://localhost:30001> (`admin` / `admin`).

Phase 2 operator commands:

```sh
make operator-test
kubectl apply -k operator/config
kubectl apply -f examples/pipeline.yaml
kubectl get pipelines -n kinetix
```

## Documentation

- [Roadmap](ROADMAP.md)
- [Tool versions](docs/tool-versions.md)
- [CI plan](docs/ci-plan.md)
- [Phase 1 local runtime plan](docs/phase-1-local-runtime-plan.md)
- [Phase 2 operator plan](docs/phase-2-operator-plan.md)
- [Phase 2 security and admission](docs/phase-2-security-and-admission.md)
- [Architecture decisions](docs/adr)

## Current Status

Phase 1 implements a hardcoded local pipeline:

```text
producer -> kinetix-input -> worker -> kinetix-output -> sink
```

Phase 2 adds the `Pipeline` CRD and a controller that reconciles worker
Deployments, ConfigMaps, and Strimzi KafkaTopic resources from a declarative
pipeline manifest.
