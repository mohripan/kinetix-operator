# Phase 1 Local Runtime Plan

Phase 1 should produce one hardcoded source-to-transform-to-sink path running
on local Kubernetes before the operator API exists.

## Proposed Shape

- Local Kubernetes: kind.
- Messaging: Strimzi-managed Kafka.
- State/cache: Redis.
- Runtime helper: Dapr sidecars where useful.
- Producer: small Go service that writes fake user events.
- Worker: small Go service that reads input records, normalizes them, emits
  output records, and exposes metrics.
- Sink reader: command or service that verifies output records.

## First Commands To Add

```sh
make kind-up
make deps-up
make docker-build
make docker-load
make deploy-demo
make smoke
make kind-down
```

## First Test Targets

- Unit tests for record parsing and transform behavior.
- Testcontainers test that verifies Kafka produce and consume behavior.
- Smoke script that sends records and verifies expected output count.

## Open Decisions Before Coding

- Final kind cluster name.
- Strimzi install method and version.
- Exact fake event schema.
- Worker topic naming convention.
- Initial metrics names.
