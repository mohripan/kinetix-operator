# ADR 0009: Phase 2 Security Baseline

## Status

Accepted

## Context

The roadmap requires security decisions before later reliability and mesh
features depend on them. Phase 2 still targets local development, but it should
define how secrets, pod security, network policy, image provenance, and supply
chain controls are expected to mature.

## Decision

Use a Kubernetes-native baseline in Phase 2:

- Operator permissions are namespace scoped and limited to `Pipeline`,
  `Deployment`, `ConfigMap`, `KafkaTopic`, `Lease`, and event operations needed
  by reconciliation.
- Pipeline author permissions are separated into `kinetix-pipeline-editor`,
  which excludes `/status` and `/finalizers`.
- The `kinetix` namespace enforces Pod Security `baseline` and audits/warns
  against `restricted` until Dapr and local dependency manifests are fully
  checked against restricted mode.
- Operator and worker pods run as non-root, use `RuntimeDefault` seccomp,
  disallow privilege escalation, drop Linux capabilities, and use read-only root
  filesystems.
- NetworkPolicies constrain worker ingress and egress paths in the local
  namespace. The operator policy restricts ingress only, because egress to the
  Kubernetes API server is cluster-specific.
- Runtime secrets are deferred to Kubernetes `Secret` references and external
  secret integration in a later phase; Phase 2 worker config contains only
  non-secret topic and bootstrap settings.
- Unsigned transform images are allowed in local development. Production-shaped
  installs must add image signature verification before Phase 4 exit.

## Consequences

Phase 2 is not production-secure, but it has concrete guardrails and names the
remaining gaps. The local install remains usable without cert-manager, an
external secrets controller, or admission image verification. Later phases can
tighten namespace policy to `restricted`, add default-deny NetworkPolicies,
enforce signed images, and replace plain config values with secret references.
