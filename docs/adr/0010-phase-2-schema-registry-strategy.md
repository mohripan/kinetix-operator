# ADR 0010: Phase 2 Schema Registry Strategy

## Status

Accepted

## Context

The `Pipeline` API already includes `spec.source.schema`, but Phase 2 does not
ship a schema registry or compatibility checker. The roadmap requires the
meaning of this field to be documented before Phase 3 lineage and compatibility
features rely on it.

## Decision

Treat `spec.source.schema` as an opaque schema subject name in Phase 2. The
controller passes the value to the worker as `KINETIX_SOURCE_SCHEMA` and does
not contact a registry. Phase 3 will introduce registry-backed subjects and
versions, transform input contracts, and compatibility checks before rollout.

## Consequences

Phase 2 manifests can carry schema intent without creating a false guarantee of
compatibility enforcement. Invalid schema names are not rejected unless they
break future structural validation. Historical lineage work must preserve the
schema string from Phase 2 records and migrate to explicit registry identity
when the registry integration lands.
