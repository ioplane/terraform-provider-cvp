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
goreleaser v2.17.1, and the Node/Python doc linters.

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

make up      # build the dev image + start the container
make shell   # shell inside it
```

## Worktree + PR loop (mandatory)

`main` is never committed to directly. Start every change in an isolated
worktree and integrate via a Pull Request (see
[`../CONTRIBUTING.md`](../CONTRIBUTING.md) and [`../AGENTS.md`](../AGENTS.md)):

```bash
scripts/worktree.sh new feat/studio/inputs-batch   # worktree under ../.worktrees/
cd ../.worktrees/feat/studio/inputs-batch
make up && make shell
# …work, make all, commit, push…
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
| Start dev container | `make up` |
| Build provider | `make build` |
| Unit tests (+race) | `make test` |
| Lint Go | `make lint` (auto-fix: `make lint-fix`) |
| Format examples | `make tffmt` / `make tffmt-check` |
| Vulnerabilities | `make vulncheck` · `make osv-scan` |
| Registry docs | `make docs` / `make docs-check` |
| Lint docs | `make lint-docs` |
| Pre-PR gate (no live CVP) | `make all` |
| Full gate (+ live acceptance) | `make verify` |
| Versions of the toolchain | `make versions` |
| Tear down | `make down` |

## Automation parity (podman-py)

Make targets, raw `podman-compose`, and the Python automation under
[`scripts/automation/`](../scripts/automation/) are three entry points to the
same container. The Python path uses
[`podman-py`](https://github.com/containers/podman-py) so CI or ad-hoc scripts
can drive the same lifecycle without the Make layer.

## Acceptance tests against the lab

Acceptance tests need a live CVP. Credentials come from `gopass`, never the
repo. See [`AGENTS.md`](../AGENTS.md) for the exact paths.

```bash
export CVP_ENDPOINT="um-cvp01.<lab-domain>:443"
export CVP_AUTH_METHOD="bearer"
export CVP_TOKEN="$(gopass show -o um-cvp/<token-path>)"
export TF_ACC=1
make testacc
```

## LSP + MCP during development

- **`gopls`** — the LSP for Go navigation, references, rename, and live
  diagnostics. Use it instead of grepping for symbols.
- **`context7` MCP** — current docs for the Plugin Framework, gRPC, and any
  library before writing code against it.
- **`arista-mcp` MCP** — authoritative Arista CVP/EOS behaviour; mandatory
  before asserting any CVP fact in code or docs (evidence discipline).
