# ADR 0003: Use OpenLineage For Lineage Events

## Status

Accepted

## Context

Lineage is a core Kinetix feature, but inventing a custom lineage model early
would make interoperability and future tooling harder. The project needs a
standard event vocabulary for jobs, runs, datasets, and related metadata.

## Decision

Use OpenLineage as the starting standard for lineage events. Kinetix-specific
record and transform metadata should be represented through compatible facets
or well-documented extensions.

## Consequences

This keeps lineage aligned with an existing ecosystem and reduces the amount of
custom protocol design required. Some record-level lineage needs may require
careful extension because OpenLineage is primarily run and dataset oriented.
