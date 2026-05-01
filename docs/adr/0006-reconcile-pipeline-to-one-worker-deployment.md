# ADR 0006: Reconcile Pipeline To One Worker Deployment First

## Status

Accepted

## Context

The initial `Pipeline` API allows a source, a list of transforms, and a sink.
The Phase 1 data plane has one worker binary that reads one Kafka input topic,
normalizes records, and writes one Kafka output topic. Supporting arbitrary
multi-transform graphs would require additional topic planning, worker chaining,
status modeling, and rollout semantics that are not proven yet.

## Decision

For Phase 2, reconcile each `Pipeline` into one worker `Deployment` using the
first transform entry. The operator also creates a worker `ConfigMap` with the
Phase 1 environment variables and records the generated resource names in
`status`.

## Consequences

This gives users the declarative `Pipeline` API and makes the hardcoded Phase 1
runtime operator-managed without overbuilding graph orchestration early. The API
can grow additively because `spec.transforms` is already a list. Until a later
phase implements multi-stage reconciliation, additional transform entries are
accepted by the type shape but not independently deployed.
