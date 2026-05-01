# Kinetix Operator Roadmap

> **For future implementation agents:** this roadmap is the product and engineering plan for Kinetix Operator. Before implementing a phase, turn that phase into a smaller task plan with explicit files, tests, commands, and commit checkpoints.

**Goal:** Build a Kubernetes-native real-time data pipeline platform with first-class lineage, observable event flow, and operator-managed infrastructure.

**Architecture:** Kinetix is split into three planes: a Kubernetes control plane, a Kafka-backed data plane, and an asynchronous lineage plane. The control plane reconciles declarative `Pipeline` resources into Kafka topics, worker deployments, config, and status. The data plane processes records through transform workers. The lineage plane consumes lineage events out of band and builds a queryable graph for debugging and auditability.

**Tech Stack:** Go, Kubernetes, Kubebuilder/controller-runtime, Kafka via Strimzi, Dapr, Redis, Postgres, OpenLineage, OpenTelemetry, Helm, Prometheus, Grafana, React or a minimal web UI, kuttl, envtest, and testcontainers.

---

## Product Vision

A data engineer should be able to define a pipeline with YAML, apply it to Kubernetes, and get a running event pipeline within seconds:

```yaml
apiVersion: pipeline.kinetix.io/v1alpha1
kind: Pipeline
metadata:
  name: user-events-enriched
spec:
  source:
    kind: KafkaSource
    topic: raw-user-events
    schema: user-event-v3
  transforms:
    - name: enrich-with-geo
      image: kinetix/geo-enricher:v1.2
      replicas: 3
    - name: filter-bots
      image: kinetix/bot-filter:v0.4
  sink:
    kind: KafkaSink
    topic: enriched-user-events
```

The platform should:

- Provision pipeline infrastructure from Kubernetes custom resources.
- Run transform workers with low-latency Kafka input and output.
- Emit lineage events for records without blocking the main data path.
- Show a live pipeline DAG with throughput, errors, health, and lineage lookup.
- Support safe rollout patterns for transform versions.
- Recover from worker crashes without data loss and with practical duplicate protection.

## Non-Goals For The First Iteration

- Matching Confluent, Materialize, Decodable, Flink, or Airflow feature depth.
- Multi-cloud support.
- A generic workflow engine.
- Perfect exactly-once semantics in Phase 1 or Phase 2.
- A polished commercial UI before the core system works.
- A custom lineage standard. Kinetix should start from OpenLineage.

## Architecture

### Control Plane

The control plane is a Kubernetes operator built with Kubebuilder and controller-runtime.

Responsibilities:

- Define the `Pipeline` API.
- Reconcile desired pipeline state into Kubernetes and Kafka resources.
- Create and update worker `Deployment` resources.
- Create config for source, transform, sink, and lineage settings.
- Track health and readiness in `status`.
- Clean up owned resources through owner references and finalizers.
- Validate pipeline manifests through admission webhooks once the API stabilizes.

### Data Plane

The data plane contains the services that process records.

Responsibilities:

- Read records from Kafka topics.
- Apply transform logic.
- Write records to downstream Kafka topics.
- Commit offsets only after successful downstream write and checkpoint update.
- Use Dapr sidecars where useful for state, pub/sub abstraction, and secrets.
- Emit metrics and traces.
- Emit lineage events asynchronously.

### Lineage Plane

The lineage plane is deliberately separate from the main record path.

Responsibilities:

- Consume lineage events from a dedicated Kafka topic.
- Store durable lineage graph data in Postgres.
- Cache hot lineage lookups in Redis.
- Answer reverse lineage questions such as "which input records produced this output record?"
- Preserve transform image, version, schema fingerprint, timestamp, and trace metadata.
- Keep the data plane moving even if lineage ingestion lags.

## Proposed Repository Layout

```text
/operator        Kubernetes operator, API types, controllers, webhooks
/workers         Worker runtime package, transform examples, test producers
/lineage         Lineage ingestion service and query API
/ui              Pipeline and lineage UI
/charts          Helm charts for local and installable deployments
/config          Kubebuilder-generated manifests and kustomize overlays
/deploy          Local cluster install manifests and environment profiles
/docs            Architecture notes, ADRs, diagrams, operating guides
/scripts         Repeatable local setup, test, demo, and cleanup scripts
/test            End-to-end fixtures and kuttl tests
```

## Milestones

### Phase 0: Project Bootstrap

**Goal:** Create a clean repository foundation before building distributed components.

Deliverables:

- Repository structure.
- Basic README with project purpose and local prerequisites.
- ADR directory and first ADRs.
- Makefile or task runner for common commands.
- Local tool version documentation.
- Initial CI plan, even if CI is not wired immediately.

Suggested ADRs:

- `docs/adr/0001-use-kubernetes-operator.md`
- `docs/adr/0002-use-kafka-as-the-event-log.md`
- `docs/adr/0003-use-openlineage-for-lineage-events.md`
- `docs/adr/0004-keep-lineage-async.md`

Exit criteria:

- A new contributor can understand what the project is and how the repo is organized.
- Tooling choices are written down before implementation starts.

### Phase 1: Local Runtime Foundations

**Goal:** Run one hardcoded pipeline end to end on Kubernetes without CRDs.

Build:

- Local Kubernetes cluster using kind or k3d.
- Strimzi-managed Kafka.
- Redis.
- Dapr.
- One fake event producer.
- One transform worker.
- One consumer or sink topic reader.
- Handwritten Helm chart or manifests for the hardcoded pipeline.
- Basic Prometheus metrics and Grafana dashboard.

Key learning targets:

- Kubernetes `Deployment`, `Service`, `ConfigMap`, `Secret`, and resource requests.
- Kafka topics, partitions, consumer groups, and offset commits.
- Dapr sidecar injection and component configuration.
- Local debugging with `kubectl`, `k9s`, `stern`, and Kafka CLI tools.

Testing:

- Go unit tests for producer and worker record handling.
- Integration tests with testcontainers for Kafka behavior where practical.
- A scripted smoke test that sends records and verifies output records.

Exit criteria:

- A developer can run a local cluster and see events flow from source to sink.
- Metrics show input count, output count, processing errors, and latency.
- The system can be torn down and recreated predictably.

### Phase 2: Operator And Pipeline CRD

**Goal:** Replace the hardcoded pipeline with a declarative `Pipeline` custom resource.

Build:

- Kubebuilder project in `/operator`.
- `Pipeline` API under `pipeline.kinetix.io/v1alpha1`.
- Reconciler that creates worker deployments and configuration.
- Kafka topic management strategy.
- `status.conditions` for readiness and failure reasons.
- Finalizers for cleanup.
- Owner references for Kubernetes-owned resources.
- Validation webhook after the first API shape is stable.

Initial `Pipeline` API shape:

```yaml
apiVersion: pipeline.kinetix.io/v1alpha1
kind: Pipeline
metadata:
  name: example
spec:
  source:
    kind: KafkaSource
    topic: input
    schema: example-v1
  transforms:
    - name: normalize
      image: kinetix/normalize:v0.1.0
      replicas: 1
  sink:
    kind: KafkaSink
    topic: output
```

Testing:

- envtest unit tests for reconciliation behavior.
- kuttl end-to-end tests for apply, update, status, and delete flows.
- Idempotency tests that run reconciliation repeatedly against the same desired state.

Exit criteria:

- `kubectl apply -f examples/pipeline.yaml` creates a working pipeline.
- `kubectl get pipelines` shows useful status.
- `kubectl delete pipeline example` cleans up owned resources.
- Reconciliation handles missing or partially-created resources.

### Phase 3: First-Class Lineage

**Goal:** Make lineage an explicit part of every processed record and every pipeline run.

Build:

- OpenLineage-based event model.
- Worker lineage emitter that publishes to a dedicated lineage topic.
- Lineage ingestion service in `/lineage`.
- Postgres schema for runs, datasets, transforms, records, and edges.
- Redis cache for recent reverse lookups.
- Query API for reverse lineage.
- Minimal UI for pipeline DAG and record lineage lookup.
- Trace correlation using OpenTelemetry.

Lineage principles:

- Lineage publishing must not block the main data path indefinitely.
- Failed lineage publishing must be visible through metrics and logs.
- Every output record should include enough metadata to find its producing transform and input records.
- Transform image, transform version, schema fingerprint, and processing timestamp must be recorded.

Testing:

- Unit tests for lineage event creation.
- Integration tests for lineage ingestion into Postgres.
- Query tests for reverse lineage traversal.
- Smoke test that sends an input record and resolves lineage from the output record.

Exit criteria:

- The UI shows a live DAG for at least one pipeline.
- A user can search for an output record ID and see the input record lineage.
- Lineage ingestion can lag without stopping the data plane.

### Phase 4: Reliability And Safe Rollouts

**Goal:** Add production-shaped reliability features without overcomplicating the earlier system.

Build:

- Dead-letter topics for failed records.
- Retry policy for transient worker failures.
- Idempotency keys and Redis-backed deduplication.
- Kafka producer idempotence and transactional patterns where appropriate.
- KEDA autoscaling from Kafka lag.
- Istio mTLS.
- Canary rollout for transform versions.
- Shadow traffic or mirrored processing for candidate transform versions.
- Output comparison tooling for canary and shadow runs.

Testing:

- Failure injection tests for worker crashes.
- Tests for duplicate input handling.
- Tests for dead-letter behavior.
- kuttl tests for rollout and rollback behavior.

Exit criteria:

- A worker crash does not lose committed records.
- Duplicate records are detected within the configured deduplication window.
- A transform version can be canaried and rolled back.
- Kafka lag can drive scaling decisions.

### Phase 5: Operations, Packaging, And Documentation

**Goal:** Make the platform installable, observable, and understandable.

Build:

- Helm chart for the operator.
- Helm chart or umbrella chart for local dependencies.
- Example pipelines.
- Runbooks for common failures.
- Grafana dashboards.
- Alerting rules.
- Architecture diagrams.
- Developer guide.
- Demo script.

Testing:

- Helm template tests.
- Install smoke test into a fresh local cluster.
- Documentation walkthrough test from an empty machine assumptions list.

Exit criteria:

- A developer can install the platform locally from documented commands.
- The demo can be run repeatedly.
- Common failure modes have runbooks.

## API Design Principles

- Keep `spec` declarative and user-owned.
- Keep `status` observed and controller-owned.
- Use Kubernetes conditions for human-readable and machine-readable state.
- Avoid embedding implementation details in the public API too early.
- Prefer additive API changes during `v1alpha1`.
- Make status useful before making the UI fancy.

## Record And Lineage Model

Each record moving through the system should have:

- Stable record ID.
- Source dataset or topic.
- Schema name and fingerprint.
- Event timestamp.
- Processing timestamp.
- Trace ID.
- Parent record IDs when produced from one or more inputs.

Each transform execution should record:

- Pipeline name and namespace.
- Transform name.
- Transform image and digest where available.
- Transform config hash.
- Input topic and output topic.
- Worker pod name.
- Success or failure status.

## Testing Strategy

Testing has to start early because distributed systems become expensive to debug quickly.

Use:

- Go unit tests for pure record, config, and API logic.
- envtest for Kubernetes controller behavior.
- testcontainers for Kafka, Redis, and Postgres integration tests.
- kuttl for Kubernetes end-to-end tests.
- Smoke scripts for local demo validation.
- Chaos or failure injection only after the basic system is stable.

Every phase should include:

- Fast tests for local development.
- At least one end-to-end confidence test.
- Clear teardown commands.
- A commit checkpoint after each working slice.

## Observability Strategy

Every service should expose:

- Request or record counts.
- Processing latency.
- Error counts by category.
- Kafka lag where relevant.
- Lineage publish success and failure counts.
- Health and readiness endpoints.

Use:

- Prometheus for metrics.
- Grafana for dashboards.
- OpenTelemetry for traces.
- Jaeger or an OpenTelemetry-compatible backend for trace inspection.
- Structured logs with pipeline name, transform name, record ID, and trace ID.

## Local Development Tooling

Expected tools:

- Docker
- kind or k3d
- kubectl
- helm
- kubebuilder
- Go
- Dapr CLI
- k9s
- stern
- kafkactl or kaf

Local scripts should eventually cover:

- Create cluster.
- Install dependencies.
- Build images.
- Load images into local cluster.
- Deploy platform.
- Apply example pipeline.
- Send sample records.
- Verify output.
- Tear down local environment.

## Decision Log

Use ADRs for decisions that shape the system. Keep them short and specific.

ADR format:

```markdown
# ADR N: Decision Title

## Status

Accepted

## Context

What problem or constraint forced this decision?

## Decision

What did we choose?

## Consequences

What gets easier, what gets harder, and what should we revisit later?
```

## Immediate Next Steps

1. Create the initial repository skeleton.
2. Add the first ADRs.
3. Write the Phase 1 local development plan with exact commands.
4. Choose kind or k3d for local Kubernetes.
5. Choose the first hardcoded demo pipeline shape.
6. Implement the smallest possible producer, worker, and sink path.
7. Add metrics before adding the operator.

## Success Criteria

This project is succeeding if:

- Each phase produces something demonstrable.
- The system remains understandable as it grows.
- Tests catch regressions before manual cluster debugging is needed.
- The operator reconciles idempotently.
- Lineage is visible, queryable, and separate from the main data path.
- Architecture decisions are documented when made.

