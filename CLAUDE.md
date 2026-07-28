# CLAUDE.md

> Guidance for automated agents and contributors working in this repository.

The canonical guide is **[`AGENTS.md`](AGENTS.md)** — read it first and follow it
exactly. This file only highlights the rules most easily missed.

## Non-negotiables

- **Work in the dev container.** No host toolchain: `task up && task shell`.
  Toolchain is defined once in `deployments/containers/Containerfile.dev`
  (`golang:1.26-trixie`).
- **Latest, pinned versions** for everything (Go 1.26.5, framework v1.19+, gRPC
  v1.82+, golangci-lint v2.12.2, latest SHA-pinned Actions).
- **`arista-mcp` MCP is mandatory** before stating any Arista CVP/EOS fact in
  code or docs. Cite it. **`context7` MCP** for library/API docs — do not rely
  on training-data recall. **`gopls`** for code navigation.
- **`gopass` for all secrets** (lab credentials, signing key). Never commit
  credentials, state, or CVP exports.
- **Verify before claiming done.** Run `task all` (and `task verify` for
  resource changes) and quote the evidence. Follow
  [`docs/testing.md`](docs/testing.md).

## Map

| Need | Go to |
|---|---|
| How to work here | [`AGENTS.md`](AGENTS.md) |
| Standards (naming/versioning/commits/changelog/Go/provider/Terragrunt) | [`docs/standards/`](docs/standards/) |
| Methodology & roles | [`docs/methodology.md`](docs/methodology.md) |
| Dev loop | [`docs/development.md`](docs/development.md) |
| Design & resources | [`design.md`](design.md) |
| Contributing & gates | [`CONTRIBUTING.md`](CONTRIBUTING.md) |

## Skills note

The global "use a skill before acting" convention still applies, but the
project's own instructions in `AGENTS.md` and `docs/standards/` **take
precedence** where they are more specific.
