# Development

Everything runs inside the Podman dev container. **No host Go, Terraform, or
Node toolchain is required.**

## Prerequisites

| Tool | Min version | Present here |
|---|---|---|
| `podman` | 5.0 | 5.8.2 |
| `podman-compose` | 1.5 | 1.5.0 |
| `podman-py` (automation) | 5.5 | 5.7.0 |
| `gopass` (secrets) | any | ✓ |

The container (`golang:1.26-trixie`, see
[`deployments/containers/Containerfile.dev`](../deployments/containers/Containerfile.dev))
bakes in: Go 1.26.5, golangci-lint v2.12.2, govulncheck, osv-scanner, gotestsum,
Terraform 1.15.8, OpenTofu 1.12.5, Terragrunt 1.1.1, tfplugindocs v0.25.0,
goreleaser v2.17.1, buf, the Node/Python doc linters, and the Astral toolchain
(uv 0.11.33 + ruff 0.16.0 + ty 0.0.64) for the podman-py automation.

## OCI / Compose specs followed

| Spec | Reference |
|---|---|
| Containerfile.5 | <https://github.com/containers/common/blob/main/docs/Containerfile.5.md> |
| containers.conf.5 | <https://github.com/containers/common/blob/main/docs/containers.conf.5.md> |
| registries.conf.5 | <https://github.com/containers/common/blob/main/docs/registries.conf.5.md> |
| containerignore.5 | <https://github.com/containers/common/blob/main/docs/containerignore.5.md> |
| Compose Specification | <https://github.com/compose-spec/compose-spec/blob/main/spec.md> |

## Bootstrap

```bash
git clone https://github.com/ioplane/terraform-provider-cvp
cd terraform-provider-cvp

task up      # build the dev image + start the container
task shell   # shell inside it
```

## Worktree + PR loop (mandatory)

`main` is never committed to directly. Start every change in an isolated
worktree and integrate via a Pull Request (see
[`../CONTRIBUTING.md`](../CONTRIBUTING.md) and [`../AGENTS.md`](../AGENTS.md)):

```bash
scripts/worktree.sh new feat/studio/inputs-batch   # worktree under ../.worktrees/
cd ../.worktrees/feat/studio/inputs-batch
task up && task shell
# …work, task all, commit, push…
scripts/worktree.sh rm feat/studio/inputs-batch    # after the PR merges
```

Optional — pin Podman to the project runtime/registry config:

```bash
export CONTAINERS_CONF=$(pwd)/deployments/containers/containers.conf
export CONTAINERS_REGISTRIES_CONF=$(pwd)/deployments/containers/registries.conf
```

## Daily loop

| Step | Command |
|---|---|
| Start dev container | `task up` |
| Build provider | `task build` |
| Unit tests (+race) | `task test` |
| Lint Go | `task lint` (auto-fix: `task lint-fix`) |
| Lint Python automation (ruff + ty) | `task lint-py` (auto-fix: `task fmt-py`) |
| Format examples | `task tffmt` / `task tffmt-check` |
| Vulnerabilities | `task vulncheck` · `task osv-scan` |
| SAST (Semgrep, host tool) | `task semgrep` |
| Registry docs | `task docs` / `task docs-check` |
| Lint docs | `task lint-docs` |
| Pre-PR gate (no live CVP) | `task all` |
| Full gate (+ live acceptance) | `task verify` |
| Versions of the toolchain | `task versions` |
| Tear down | `task down` |

## Automation parity (podman-py)

`task` targets, raw `podman-compose`, and the Python automation under
[`scripts/automation/`](../scripts/automation/) are three entry points to the
same container. The Python path uses
[`podman-py`](https://github.com/containers/podman-py) (the library, over the
Podman REST socket) so CI or ad-hoc scripts can drive the same image build and
container lifecycle without the `task` layer. It is linted with **ruff** and
type-checked with **ty**, run via **uv** (`task lint-py` / `task fmt-py`).

## Acceptance tests against the lab

Acceptance tests need a live CVP. Credentials come from `gopass`, never the
repo. See [`AGENTS.md`](../AGENTS.md) for the exact paths.

```bash
export CVP_ENDPOINT="um-cvp01.<lab-domain>:443"
export CVP_AUTH_METHOD="bearer"
export CVP_TOKEN="$(gopass show -o um-cvp/<token-path>)"
export TF_ACC=1
task testacc
```

## LSP + MCP during development

- **`gopls`** — the LSP for Go navigation, references, rename, and live
  diagnostics. Use it instead of grepping for symbols.
- **`context7` MCP** — current docs for the Plugin Framework, gRPC, and any
  library before writing code against it.
- **`arista-mcp` MCP** — authoritative Arista CVP/EOS behaviour; mandatory
  before asserting any CVP fact in code or docs (evidence discipline).
