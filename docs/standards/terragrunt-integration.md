# Terragrunt integration

How consumers drive `terraform-provider-cvp` with
[Terragrunt](https://terragrunt.gruntwork.io/) (pinned **v1.1.1** in the dev
container, on the 1.x backward-compatibility commitment). Terragrunt is the
expected orchestration layer for multi-workspace CVP change management.

## Why Terragrunt here

CVP's apply model is workspace-scoped: one `cvp_workspace` is the root of a
dependency chain (studio inputs, tags, change control all reference its id), and
the provider deliberately does **not** support multiple workspaces in one apply
(`design.md` D2). That maps cleanly onto Terragrunt's **unit** = one workspace,
composed into a **stack**.

## Version baseline (2026)

- **Terragrunt 1.0** (2026-03-30) froze the CLI/HCL contract: `run`, `exec`,
  `find`, `list`; the unified `--filter` system; structured run reports; and
  automatic provider caching with OpenTofu 1.10+.
- **Terragrunt 1.1** (July 2026) adds **stack dependencies** (declare
  unit-to-unit wiring in `terragrunt.stack.hcl`), the redesigned **catalog** as
  default, and read-based **change detection**.

Use `unit` / `stack` terminology (not the legacy "module"), and the `run`
subcommand (not the deprecated bare `terragrunt apply`).

## Recommended layout

```text
live/
├── root.hcl                         # remote_state + provider generate + common inputs
├── s1/
│   └── cvp/
│       ├── bgp-as-bump/
│       │   └── terragrunt.hcl        # unit → one cvp_workspace + its inputs + CC
│       └── evpn-rollout/
│           └── terragrunt.hcl
└── _catalog/                         # reusable unit templates (redesigned catalog)
```

## Provider config generation

Generate the provider block from Terragrunt so credentials come from the
environment (sourced from `gopass`), never from committed HCL:

```hcl
# root.hcl
generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite"
  contents  = <<-EOF
    terraform {
      required_providers {
        cvp = {
          source  = "ioplane/cvp"
          version = "~> 0.1"
        }
      }
    }
    provider "cvp" {
      endpoint    = "${get_env("CVP_ENDPOINT")}"
      auth_method = "bearer"
      token       = "${get_env("CVP_TOKEN")}"
    }
  EOF
}
```

- Pin the provider `version` with a pessimistic `~>` constraint; while `0.x`,
  pin the minor (`~> 0.2`) because minors may break (see
  [`versioning.md`](versioning.md)).
- Commit the generated `.terraform.lock.hcl` for reproducible provider
  selection.

## Dependencies between units

Prefer **stack dependencies** (1.1) to wire, e.g., a change-control unit after
the workspace-inputs unit, instead of threading `dependency` output paths by
hand. Keep `mock_outputs` for `plan` on a not-yet-applied dependency.

## CI / change detection

Use read-based change detection (`terragrunt find --filter`/run queue) so a CI
run only touches units whose inputs actually changed. Do not auto-approve CVP
change control from CI — the provider defaults `auto_approve = false`; human
approval stays a distinct, gated step (`design.md` non-goal: "do not automate
approval").

## Examples

Worked HCL + Terragrunt examples live under `examples/` and are exercised by the
`terraform` CI job (`terraform fmt -check`).
