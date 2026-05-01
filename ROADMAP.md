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
- Make tenancy, security boundaries, schema compatibility, retention, and
  operational resource expectations explicit before production-shaped features
  depend on them.

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
- Enforce admission-time policy for namespace ownership, schema references,
  allowed images, and resource guardrails as those policies become available.

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
- Surface Kafka lag, retry pressure, dead-letter counts, and throttling signals
  so backpressure decisions are visible instead of implicit.

### Lineage Plane

The lineage plane is deliberately separate from the main record path.

Responsibilities:

- Consume lineage events from a dedicated Kafka topic.
- Store durable lineage graph data in Postgres.
- Cache hot lineage lookups in Redis.
- Answer reverse lineage questions such as "which input records produced this output record?"
- Preserve transform image, version, schema fingerprint, timestamp, and trace metadata.
- Keep the data plane moving even if lineage ingestion lags.

### Security And Supply Chain

Security is a first-class architecture concern, not only a Phase 4 mesh feature.
The roadmap should progressively answer:

- How runtime secrets are sourced, rotated, and mounted.
- Which namespaces and service accounts can create or reconcile `Pipeline`
  resources.
- Which pod security standards and network policies are required for local and
  production profiles.
- Whether transform images must be signed and how signatures are verified.
- How SBOMs, vulnerability scanning, and base image policies fit into release
  workflows.
- How Kubernetes RBAC, admission control, Istio policies, and network policies
  complement each other instead of duplicating responsibility.

### Tenancy Model

Kinetix needs an explicit tenancy decision before production design hardens. The
first iteration may choose single-tenant operation, but the choice must be
documented because it shapes namespace layout, Kafka topic naming, resource
quotas, lineage isolation, RBAC, and noisy-neighbor protection. If multi-tenancy
is deferred, the API and resource naming should avoid choices that make it
expensive to add later.

### Schema And Compatibility

The record model includes schema names and fingerprints, but event systems also
need schema compatibility rules. Kinetix should define a schema registry
strategy, how transforms declare accepted schema versions, what compatibility
modes are supported, and how invalid or unexpected records are handled. Schema
failures should be observable and should integrate with retry and dead-letter
behavior.

### Backpressure And Flow Control

Kafka buffering and KEDA autoscaling are not a full backpressure strategy.
Kinetix must decide what happens when input rate exceeds transform capacity:
scale, throttle, pause, shed, route to dead-letter topics, or accept lag. Those
decisions should be visible through metrics, status, alerts, and runbooks.

### Disaster Recovery, Retention, And Cost Model

The roadmap should define practical RPO/RTO targets for local and production
profiles, data retention for Kafka topics and lineage storage, backup and
restore expectations for Postgres, and resource sizing guidance for Kafka,
workers, Redis, and Postgres. These do not need exact cloud prices, but they do
need enough numbers and defaults for operators to reason about requests,
limits, connection pools, partitions, and storage growth.

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
- Initial RBAC design for operator permissions and user permissions to create
  `Pipeline` resources in approved namespaces.
- Initial tenancy ADR that states whether Phase 2 targets single-tenant,
  namespace-per-team, or cluster-wide multi-tenant operation.
- Initial security ADR covering secrets management, pod security, network
  policy, image signing, and supply-chain posture for early phases.
- Initial schema registry ADR that records the planned registry integration,
  compatibility mode, and how `spec.source.schema` is interpreted before the
  registry exists.

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
- RBAC smoke tests that prove a user can create allowed `Pipeline` resources
  and cannot mutate controller-owned status or resources directly.

Exit criteria:

- `kubectl apply -f examples/pipeline.yaml` creates a working pipeline.
- `kubectl get pipelines` shows useful status.
- `kubectl delete pipeline example` cleans up owned resources.
- Reconciliation handles missing or partially-created resources.
- The tenancy, security baseline, and schema registry strategy are captured in
  ADRs, even if their full implementation lands in later phases.

### Phase 3: First-Class Lineage

**Goal:** Make lineage an explicit part of every processed record and every pipeline run.

Build:

- OpenLineage-based event model.
- Schema registry integration for source, transform, and sink schemas.
- Transform schema contracts that declare accepted input versions and emitted
  output versions.
- Compatibility checks for `Pipeline` manifests before workers are rolled out.
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
- Schema compatibility failures must be visible as data quality events and must
  not be confused with infrastructure failures.

Testing:

- Unit tests for lineage event creation.
- Unit tests for schema compatibility decisions.
- Integration tests for lineage ingestion into Postgres.
- Integration tests for schema registry lookup and incompatible schema rejection.
- Query tests for reverse lineage traversal.
- Smoke test that sends an input record and resolves lineage from the output record.

Exit criteria:

- The UI shows a live DAG for at least one pipeline.
- A user can search for an output record ID and see the input record lineage.
- Lineage ingestion can lag without stopping the data plane.
- A transform receiving an unsupported schema version is handled predictably and
  exposes a clear status, metric, and event.

### Phase 4: Reliability And Safe Rollouts

**Goal:** Add production-shaped reliability features without overcomplicating the earlier system.

Build:

- Dead-letter topics for failed records.
- Retry policy for transient worker failures.
- Idempotency keys and Redis-backed deduplication.
- Kafka producer idempotence and transactional patterns where appropriate.
- Explicit backpressure policy for slow transforms, including lag thresholds,
  throttling behavior, worker pause/resume behavior, and when records are routed
  to dead-letter topics.
- KEDA autoscaling from Kafka lag.
- Istio mTLS.
- Kubernetes NetworkPolicies for operator, worker, Kafka, Redis, Postgres, and
  UI communication paths.
- Pod security standards for operator-managed workloads.
- Secrets management implementation using the strategy chosen in the security
  ADR.
- Image signature verification for operator and transform images.
- Canary rollout for transform versions.
- Shadow traffic or mirrored processing for candidate transform versions.
- Output comparison tooling for canary and shadow runs.

Testing:

- Failure injection tests for worker crashes.
- Tests for duplicate input handling.
- Tests for dead-letter behavior.
- Tests for slow transform backpressure and lag-driven behavior.
- Tests that NetworkPolicies block unexpected traffic and allow required paths.
- Tests that unsigned or untrusted images are rejected when verification is enabled.
- kuttl tests for rollout and rollback behavior.

Exit criteria:

- A worker crash does not lose committed records.
- Duplicate records are detected within the configured deduplication window.
- A transform version can be canaried and rolled back.
- Kafka lag can drive scaling decisions.
- Slow consumers have documented and tested behavior before lag becomes
  unbounded.
- The security baseline covers mTLS, RBAC, network policy, pod security,
  secrets, and image verification.

### Phase 5: Operations, Packaging, And Documentation

**Goal:** Make the platform installable, observable, and understandable.

Build:

- Helm chart for the operator.
- Helm chart or umbrella chart for local dependencies.
- Example pipelines.
- Runbooks for common failures.
- Backup and restore runbooks for Postgres lineage data and any operator-owned
  durable state.
- Retention configuration for Kafka topics, dead-letter topics, lineage tables,
  and hot Redis caches.
- RPO/RTO targets for local-demo and production-shaped profiles.
- Resource sizing guide for Kafka partitions, worker requests and limits, Redis
  memory, Postgres storage, and Postgres connection pooling.
- SBOM generation and vulnerability scanning in the release process.
- Signed image release workflow for operator and example worker images.
- Grafana dashboards.
- Alerting rules.
- Architecture diagrams.
- Developer guide.
- Demo script.

Testing:

- Helm template tests.
- Install smoke test into a fresh local cluster.
- Documentation walkthrough test from an empty machine assumptions list.
- Restore drill for lineage Postgres from a backup fixture.
- Retention test or documented verification for Kafka and lineage cleanup.

Exit criteria:

- A developer can install the platform locally from documented commands.
- The demo can be run repeatedly.
- Common failure modes have runbooks.
- Operators can estimate baseline resource needs and storage growth for a
  pipeline before installing it.
- Backup, restore, retention, and signed release workflows are documented and
  tested at least once.

## API Design Principles

- Keep `spec` declarative and user-owned.
- Keep `status` observed and controller-owned.
- Use Kubernetes conditions for human-readable and machine-readable state.
- Avoid embedding implementation details in the public API too early.
- Prefer additive API changes during `v1alpha1`.
- Make status useful before making the UI fancy.
- Treat tenancy, schema references, resource limits, and security policy as API
  design constraints rather than operational afterthoughts.

## Record And Lineage Model

Each record moving through the system should have:

- Stable record ID.
- Source dataset or topic.
- Schema name and fingerprint.
- Event timestamp.
- Processing timestamp.
- Trace ID.
- Parent record IDs when produced from one or more inputs.
- Schema registry subject or equivalent schema identity when registry
  integration exists.

Each transform execution should record:

- Pipeline name and namespace.
- Transform name.
- Transform image and digest where available.
- Transform config hash.
- Input topic and output topic.
- Worker pod name.
- Success or failure status.
- Accepted input schema version and emitted output schema version.
- Backpressure state, retry count, and dead-letter routing decision when
  applicable.

## Tenancy And Security Strategy

Early ADRs should answer:

- Is the first install model single-tenant, namespace-per-team, or shared
  cluster multi-tenant?
- Which Kubernetes subjects can create `Pipeline` resources?
- Which service accounts does the operator use, and what resources can they
  mutate?
- Where do Kafka, Redis, Postgres, registry, and worker secrets come from?
- Which pod security profile applies to operator-managed workloads?
- Which network paths are allowed before and after Istio is installed?
- Are unsigned transform images allowed in local development only, or also in
  production-shaped profiles?

The implementation can start permissive for local development, but the intended
production posture must be explicit.

## Schema Evolution Strategy

Schema compatibility must be part of the event platform contract:

- Source and sink schemas should eventually refer to registry subjects and
  versions, not just free-form strings.
- Transforms should declare compatible input schema versions and produced output
  schema versions.
- Admission or reconciliation should reject known-incompatible pipeline graphs.
- Runtime schema mismatches should produce clear metrics, events, and
  dead-letter records.
- Lineage records should preserve schema identity so historical outputs remain
  explainable after schema evolution.

## Backpressure Strategy

Backpressure behavior should be designed before reliability features harden:

- Track Kafka lag by pipeline, transform, and consumer group.
- Define thresholds for warning, scaling, pausing, and failure states.
- Decide when to scale workers and when scaling is unsafe.
- Decide whether producers are slowed, workers pause consumption, records are
  dropped, or records move to dead-letter topics.
- Expose backpressure state through metrics, logs, Kubernetes events, and
  `Pipeline` status.

## Operations Strategy

Production-shaped operations require documented defaults and drills:

- Retention periods for Kafka, dead-letter topics, lineage tables, and cache
  entries.
- Backup cadence and restore procedure for Postgres lineage data.
- RPO/RTO targets by environment profile.
- Resource sizing guidance for common pipeline shapes.
- Release artifacts with SBOMs, vulnerability scan results, and signed images.

These details can mature over phases, but they should be tracked as roadmap
requirements rather than discovered during packaging.

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
- Backpressure state and lag threshold transitions.
- Schema compatibility failures and dead-letter routing counts.
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
