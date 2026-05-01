# ADR 0008: Use Single Tenant Namespace For Phase 2

## Status

Accepted

## Context

Phase 2 needs a concrete tenancy boundary before RBAC, topic ownership, status
updates, and worker resources harden. The roadmap allows the first iteration to
choose single-tenant operation if the decision is explicit and does not block a
future namespace-per-team model.

## Decision

Run Phase 2 as a single tenant install in the `kinetix` namespace. Pipeline
authors receive namespace-scoped permissions to create and update `Pipeline`
resources, while the operator service account owns reconciliation of child
resources and status. Resource names include the pipeline name but do not encode
team or tenant identity yet.

## Consequences

The initial RBAC and NetworkPolicy surface stays small enough to test locally.
Users cannot directly mutate `Pipeline` status through the supplied
`kinetix-pipeline-editor` role. Future multi-tenant work should move to
namespace-per-team isolation before introducing shared-cluster tenant routing,
topic prefixes, quota policy, or lineage isolation.
