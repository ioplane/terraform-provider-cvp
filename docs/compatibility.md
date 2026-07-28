<div align="center">

# Compatibility

</div>

The consumer-facing compatibility contract for `terraform-provider-cvp`: which
Terraform / OpenTofu versions each capability needs, the wire protocol, and the
toolchain the provider is built against. Version-support changes follow
[SemVer](standards/versioning.md) (raising a minimum is a MINOR bump pre-1.0, a
breaking change post-1.0).

## Wire protocol & toolchain

| Item | Value |
|---|---|
| Plugin protocol | **6.0** (`terraform-registry-manifest.json`) |
| Framework | `terraform-plugin-framework` **v1.19.0** |
| Go (build) | **1.26** |
| Lab validated against | the netlab2 CloudVision Portal lab |

The provider is a single protocol-6 server; every capability below travels that
one protocol. A newer _core_ is required only for the capability that uses it —
the base resources work on any protocol-6 core.

## Feature → minimum core version

| Capability | Surface | Terraform | OpenTofu | Shipped |
|---|---|---|---|---|
| Managed resources + import | `cvp_workspace`, `cvp_studio_inputs` | any protocol-6 (**≥ 1.0**) | **≥ 1.6** | ✅ |
| Provider-defined functions | `provider::cvp::studio_path` | **≥ 1.8** | **≥ 1.7** | ✅ |
| **Actions** | `cvp_workspace_build` / `_submit` / `_abandon` / `_cancel_build` / `_rollback` / `_rebase` | **≥ 1.14** | **not yet supported** | ✅ |
| Write-only attributes | studio secret inputs (D6) | ≥ 1.11 | ≥ 1.11 | ⏳ planned |
| Ephemeral resources | `cvp_service_account_token` (backlog) | ≥ 1.10 | ≥ 1.11 | ⏳ planned |
| Managed resource identity | server-minted ids (backlog) | ≥ 1.12 | — | ⏳ planned |

> [!IMPORTANT]
> **Actions are Terraform-only.** They require Terraform **≥ 1.14**; OpenTofu
> (through 1.12) does not implement `lifecycle.action_trigger`. The
> `cvp_workspace` **resource** and the `studio_path` **function** work on
> OpenTofu — only the workflow _actions_ do not. Acceptance for actions runs on
> Terraform 1.15.8 in the dev container.

## Provider ↔ CVP

The provider talks to CVP's Resource APIs over gRPC/TLS (protocol-6 core ↔
provider ↔ CVP). It targets the modern `arista.*.v1` Resource APIs
(`workspace.v1`, `studio.v1`; `changecontrol.v1` planned) and is validated
continuously against the netlab2 lab; a formal CVP-release support matrix lands
with the verification phase (`design.md` D8, `methodology.md`).

## Declaring the requirement in config

```hcl
terraform {
  # Raise to >= 1.14 only if you use the cvp_workspace_* actions.
  required_version = ">= 1.8.0"
  required_providers {
    cvp = {
      source  = "ioplane/cvp"
      version = "~> 0.2"
    }
  }
}
```
