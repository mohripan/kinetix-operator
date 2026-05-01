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

## Implemented Decisions

- kind cluster name: `kinetix`.
- kind node image: `kindest/node:v1.32.5` to match the Kubernetes `1.32.x` tool baseline and Strimzi `0.45.0`.
- Kubernetes namespace: `kinetix`.
- Strimzi install method: Helm chart `strimzi/strimzi-kafka-operator` version `0.45.0`.
- Dapr install method: Helm chart `dapr/dapr` version `1.15.8`.
- Kafka cluster name: `kinetix-kafka`.
- Input topic: `kinetix-input`.
- Output topic: `kinetix-output`.
- Fake event schema: JSON `UserEvent` with `id`, `user_id`, `event_type`, `timestamp`, and optional string attributes.
- Normalized output schema: JSON record with `parent_id`, source topic, schema fingerprint, and processing timestamp.
- Initial metrics:
  - `kinetix_input_records_total`
  - `kinetix_output_records_total`
  - `kinetix_processing_errors_total`
  - `kinetix_processing_latency_seconds`

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

## Runtime URLs

After `make deploy-demo`, local NodePorts expose:

- Prometheus: <http://localhost:30000>
- Grafana: <http://localhost:30001> with `admin` / `admin`

## First Test Targets

- Unit tests for record parsing and transform behavior.
- Testcontainers test that verifies Kafka produce and consume behavior once dependency downloads are available in the development environment.
- Smoke script that sends records and verifies expected output count.
