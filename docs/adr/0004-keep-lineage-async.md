# ADR 0004: Keep Lineage Async

## Status

Accepted

## Context

Lineage should make pipelines easier to debug and audit, but it must not become
the bottleneck that stops the main record path. The data plane needs to keep
processing when lineage ingestion is slow or temporarily unavailable.

## Decision

Publish lineage events asynchronously to a dedicated Kafka topic. A separate
lineage service will consume those events and write durable graph data to
Postgres, with Redis used later for hot lookup caching.

## Consequences

The main data path can continue while lineage ingestion lags, and lineage
failures can be tracked through metrics and logs. The tradeoff is eventual
consistency: recent records may not be queryable immediately, and the system
must expose lag and publish failures clearly.
