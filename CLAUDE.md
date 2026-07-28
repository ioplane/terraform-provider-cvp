<div align="center">

# CLAUDE.md

</div>

> Guidance for automated agents and contributors working in this repository.

The canonical guide is **[`AGENTS.md`](AGENTS.md)** — read it first and follow it
exactly. This file only highlights the rules most easily missed.

## Non-negotiables

- **Work in the dev container.** No host toolchain: `task up && task shell`.
  Toolchain is defined once in `deployments/containers/Containerfile.dev`
  (`golang:1.26-trixie`).
- **Latest, pinned versions** for everything (Go 1.26.5, framework v1.19+, gRPC
  v1.82+, golangci-lint v2.12.2, latest SHA-pinned Actions).
- **Source of truth — never from memory.** Before stating or building on any
  CVP/EOS or library API fact (endpoint, RPC/field, TerminAttr flag, EOS CLI,
  port, auth/RBAC), verify it against a source: **`arista-mcp` MCP** (CVP/EOS
  behavior, mandatory + cite), **`cvprac`** (`../cvprac/cvprac/cvp_api.py` — the
  current CVP REST endpoint/body), the **cloudvision-apis proto** (gRPC fields),
  **`arista-cvp-re`** (auth/RBAC internals), **`context7` MCP** (other libs).
  Never trust training-data recall — a guessed endpoint's 4xx is not evidence
  about permissions. See the `cvp-api-lookup` skill,
  `~/.claude/rules/cvp-eos-sources.md`, and AGENTS.md. **`gopls`** for code nav.
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
