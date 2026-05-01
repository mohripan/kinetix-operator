# ADR 0001: Use A Kubernetes Operator

## Status

Accepted

## Context

Kinetix needs to turn declarative pipeline definitions into Kubernetes
resources, runtime configuration, and observed status. The platform also needs
idempotent reconciliation, ownership, cleanup, and eventually validation.

## Decision

Build the control plane as a Kubernetes operator using Kubebuilder and
controller-runtime. The operator will reconcile `Pipeline` custom resources into
worker deployments, configuration, Kafka resources, and status conditions.

## Consequences

This makes the user-facing API Kubernetes-native and gives us established
patterns for reconciliation, owner references, finalizers, status, and webhooks.
It also means contributors must understand Kubernetes controller behavior, and
local development will require envtest and disposable clusters.
