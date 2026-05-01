# ADR 0002: Use Kafka As The Event Log

## Status

Accepted

## Context

The data plane needs durable ordered event transport, partitioned scale,
consumer groups, replay, and practical local development support. Kinetix also
needs topic-level boundaries between source, transform, sink, and lineage flow.

## Decision

Use Kafka as the primary event log for pipeline records. In Kubernetes, start
with Strimzi-managed Kafka for local and installable environments.

## Consequences

Kafka provides a familiar durability and replay model for streaming pipelines
and gives workers a clear offset and consumer-group contract. It adds
operational complexity, so Phase 1 must make local setup and teardown
repeatable before the operator starts managing pipeline resources.
