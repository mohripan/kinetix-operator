# ADR 0007: Manage Kafka Topics With Strimzi KafkaTopic Resources

## Status

Accepted

## Context

Kinetix uses Strimzi-managed Kafka for the local runtime. The operator needs a
Kafka topic management strategy for source and sink topics. It could call the
Kafka Admin API directly, shell out to Kafka tooling, or delegate topic
reconciliation to Strimzi using `KafkaTopic` custom resources.

## Decision

Manage Phase 2 topics by reconciling Strimzi `KafkaTopic` resources owned by
the `Pipeline`. The first implementation creates source and sink topic resources
with conservative defaults for partitions, replicas, and retention.

## Consequences

Topic lifecycle stays Kubernetes-native and can be observed with `kubectl`
alongside the rest of the pipeline resources. The Kinetix operator avoids
embedding Kafka admin credentials and direct broker administration logic at this
stage. This couples local topic management to Strimzi, so future non-Strimzi
install profiles will need a separate topic management strategy or an
abstraction in the controller.
