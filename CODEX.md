# CODEX.md

This repository's working guide for agents is **[`AGENTS.md`](AGENTS.md)** —
read it first. It is the single source of truth for role, environment,
standards, tooling (LSP + `context7`/`arista-mcp` MCP), the netlab2 CVP lab,
`gopass` secrets, workflow, and quality gates.

## Current context

- Repository: `ioplane/terraform-provider-cvp`
- State: **v0.1 skeleton** — provider schema authored; CRUD stubbed with
  `TODO(P1)`; three resources (`cvp_workspace`, `cvp_studio_inputs`,
  `cvp_change_control`).
- Stack: Go 1.26.5, `terraform-plugin-framework` v1.19+, protocol 6, gRPC/TLS.
- Standards: SemVer 2.0.0 · Conventional Commits 1.0.0 · Keep a Changelog 1.1.0,
  detailed under [`docs/standards/`](docs/standards/).
- Methodology: gated-iterative delivery — [`docs/methodology.md`](docs/methodology.md).
- Dev loop: Podman + `podman-compose`, `golang:1.26-trixie` — `task up && task shell`.

Future work starts from the skeleton and the design in [`design.md`](design.md).
Do not assert CVP behaviour without corroborating it via `arista-mcp` and, where
possible, the live lab. Never commit secrets — they live in `gopass`.
