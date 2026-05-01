# Tool Versions

This project uses explicit local tool versions so contributors can reproduce
cluster and controller behavior during development.

## Baseline

| Tool | Target version | Notes |
| --- | --- | --- |
| Go | 1.24.x | Use the latest patch release in the 1.24 line unless a module requires otherwise. |
| Docker | 26.x or newer | Required for local images, kind, and testcontainers. |
| kind | 0.27.x | Preferred local Kubernetes runtime for Phase 1 unless k3d proves necessary. |
| kubectl | 1.32.x | Match the local cluster minor version when possible. |
| Helm | 3.17.x | Used for Strimzi, Dapr, and project charts. |
| Kubebuilder | 4.x | Introduced in Phase 2 for operator scaffolding. |
| Dapr CLI | 1.15.x | Used when Dapr sidecars/components are added. |
| k9s | 0.32.x or newer | Optional but recommended for local debugging. |
| stern | 1.31.x or newer | Optional but recommended for multi-pod logs. |
| kafkactl or kaf | Current stable | Used for inspecting Kafka topics and records locally. |
| kuttl | 0.19.x | Used for Kubernetes end-to-end tests in later phases. |

## Version Management

- Prefer installing tools through repeatable package managers such as
  `asdf`, `mise`, `brew`, `winget`, or `choco`.
- Keep this file updated when a phase introduces a new required tool.
- When a tool is only needed for optional debugging, mark it as optional rather
  than blocking local setup on it.

## GoLand Notes

- Open the repository root in GoLand.
- Enable Go modules once Phase 1 or Phase 2 creates the first `go.mod`.
- Use the Makefile targets as run configurations for common tasks.
- Keep Kubernetes contexts pointed at the local kind cluster while developing
  controller and manifest behavior.
