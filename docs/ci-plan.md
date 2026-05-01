# CI Plan

CI is not wired in Phase 0, but the expected pipeline is documented now so early
implementation work has clear quality gates.

## Initial Checks

1. Documentation presence check using `make docs`.
2. Go formatting with `gofmt` once Go packages exist.
3. Go unit tests with `go test ./...`.
4. Linting with `golangci-lint` once the first Go module is added.
5. Helm template validation once charts exist.

## Later Checks

1. Controller envtest suite for reconciler behavior.
2. Testcontainers-backed integration tests for Kafka, Redis, and Postgres
   behavior that can run outside Kubernetes.
3. kuttl tests against a disposable kind cluster for CRD and controller flows.
4. Smoke tests for the local demo path.
5. Container image build checks for operator, workers, and lineage services.

## Proposed GitHub Actions Jobs

| Job | Trigger | Purpose |
| --- | --- | --- |
| `docs` | pull request, main | Verify required docs and ADRs exist. |
| `go-test` | pull request, main | Format and test all Go modules. |
| `charts` | pull request, main | Render Helm charts and validate Kubernetes manifests. |
| `integration` | pull request, main | Run testcontainers integration tests. |
| `kind-smoke` | main, manual | Run local Kubernetes smoke tests. |

## Principles

- Keep fast checks on every pull request.
- Put cluster-heavy tests behind explicit jobs until they are stable.
- Publish logs and rendered manifests as artifacts for failed Kubernetes jobs.
- Prefer deterministic teardown over reusing mutable CI clusters.
