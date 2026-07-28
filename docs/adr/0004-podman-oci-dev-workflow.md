# ADR 0004 — Podman/OCI dev, build, and CI workflow

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

The build/test/lint/release loop should be reproducible across hosts and CI
without requiring a host Go/Terraform/Node toolchain, and should follow open
standards rather than a single vendor's tooling.

## Decision

- Development, build, and test run inside a **Podman** container built from
  `golang:1.26-trixie` (`deployments/containers/Containerfile.dev`), orchestrated
  with **`podman-compose`** (Compose Specification) and automatable via
  **`podman-py`**.
- Follow the **OCI / containers ecosystem specs**: Containerfile.5,
  containers.conf.5, registries.conf.5, containerignore.5, and the
  compose-spec.
- CI uses the same `golang:1.26-trixie` image; GitHub Actions are pinned by
  commit SHA; releases go through goreleaser v2.

## Consequences

- One toolchain definition (`Containerfile.dev`) is the single source of truth
  for tool versions; `make`, raw `podman-compose`, and `scripts/automation/`
  (podman-py) are three entry points to it.
- No host Go required; contributor onboarding is `task up && task shell`.
- Rootless-friendly, daemonless, and vendor-neutral (Docker-compatible via the
  compose-spec if a contributor prefers it).
- Base image floats within the `1.26` minor for patch pickup; exact tool
  versions are pinned as build `ARG`s.
