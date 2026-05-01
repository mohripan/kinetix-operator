# ADR 0005: Use Controller Runtime For The Phase 2 Operator

## Status

Accepted

## Context

Phase 2 introduces the first working `Pipeline` controller. The roadmap calls
for a Kubebuilder project, but the local environment does not always have the
`kubebuilder` binary installed. We still need the generated project shape,
controller conventions, reconciliation APIs, status updates, ownership, and
test support that Kubebuilder projects normally provide.

## Decision

Build the operator as a Go module in `/operator` using controller-runtime
directly, with a Kubebuilder-compatible layout for API types, controllers,
config, RBAC, CRDs, and manager entrypoint. Keep the module on the Kubernetes
1.32-compatible dependency line used by the Phase 1 local cluster target.

## Consequences

The operator remains compatible with the Kubebuilder ecosystem without making
the scaffold command a prerequisite for this repository state. Contributors can
run normal Go tests immediately, and optional envtest coverage can be enabled
when Kubebuilder test assets are installed. The cost is that generated markers
and CRD updates are manual until code generation is wired into the build.
