<div align="center">

# Backlog — modern Terraform / Terragrunt / enterprise capabilities

</div>

A living, prioritized list of capabilities to evaluate **beyond** the v0.2 CRUD
scope (`design.md`). It captures modern Terraform (1.10–1.15), Terragrunt (1.1),
and HCP Terraform / Terraform Enterprise features that fit CVP, plus patterns
learned from similar providers. Items graduate into an [ADR](adr/) and a sprint
via the methodology's Definition of Ready ([`methodology.md`](methodology.md)).

Priority: **P0** (fold into current v0.2/v0.3 design) · **P1** (v0.x, high value)
· **P2** (v1.x / opportunistic).

## 1. Terraform Plugin Framework capabilities

| Capability | TF min | CVP fit | Prio |
|---|---|---|---|
| **Actions** (`ProviderWithActions`) | 1.14 | ✅ **Done for workspace** (ADR 0006): `cvp_workspace_build / _cancel_build / _submit / _abandon / _rollback / _rebase` replace the `auto_*` booleans. Change-control `start / approve / rollback` actions remain to do. See §6. | **P0** |
| **Managed resource identity** (`ResourceWithIdentity`) | 1.12 | Every CVP object has a stable server-minted id (`workspace_id`, studio id, cc id). Expose it as the resource **identity** → import-by-identity, cross-run stable identification, cleaner drift. | **P1** |
| **Write-only attributes** (`WriteOnly: true`) | 1.11 | Studio `is_secret` inputs and any password/token argument (`design.md` D6). Never persisted to plan/state. Pair with a `*_wo_version` rotation trigger (the `aws_db_instance` pattern; also `terraform-provider-pgsteward`). | **P0** |
| **Ephemeral resources** (`EphemeralResource`, Open/Renew/Close) | 1.10 | `cvp_service_account_token` / `cvp_api_token` — mint a short-lived CVP token (via `serviceaccount.v1.TokenConfigService`) for other providers/tools **without** writing it to state; `RenewAt` for TTL tokens. Mirrors `pgsteward_cloud_token`, `hashicorp/vault`, `Azure/azapi`. | **P1** |
| **Provider-defined functions** | 1.8 | ✅ **`provider::cvp::studio_path(...)` shipped** — assembles a `cvp_studio_inputs` path (bracket notation) from ordered segments; generic, pure, offline. Still open: `provider::cvp::merge_inputs(...)` for `inputs_json` composition. | **P2** |
| **Deferred actions** (experimental) | 1.9+ (exp) | Defer a resource's create when `workspace_id` (or another dependency) is unknown at plan, instead of erroring — smoother `-target`-free multi-stage applies. Experimental; track before committing. | **P2** |
| **`terraform test` (`.tftest.hcl`)** | 1.6 | Native config-driven tests alongside the Go acceptance suite (`docs/testing.md`) — cheap plan/apply assertions runnable without the Go toolchain. | **P1** |
| **`check` blocks / continuous validation** | 1.5 | Post-apply assertions on CVP state (e.g., workspace build succeeded, change-control reached `SUCCESS`, no compliance drift) surfaced as warnings, not apply failures. | **P2** |
| **Config validators / plan modifiers / state & identity upgraders** | 1.x | Already mandated by the per-resource Definition of Done (`docs/standards/terraform-provider-best-practices.md`). Listed for completeness. | **P0** |

## 2. Authentication & credentials

| Capability | CVP fit | Prio |
|---|---|---|
| **`TokenSource` abstraction** | Factor auth behind a `Token(ctx) (string, error)` interface (like `terraform-provider-pgsteward`): `static` (today) → `oidc` / workload-identity that mints a CVP service-account token out-of-band. Enables short-lived tokens with automatic refresh. | **P1** |
| **HCP Terraform dynamic provider credentials (OIDC)** | In HCP Terraform / TFE, exchange the run's OIDC token for a CVP service-account token — no static `token` stored in the workspace. Requires the `TokenSource` above + a documented OIDC trust setup. | **P2** |

## 3. HCP Terraform / Terraform Enterprise (enterprise)

| Capability | CVP fit | Prio |
|---|---|---|
| **Policy as code (Sentinel / OPA)** | Ship example policies: forbid `auto_approve = true` outside non-prod, enforce the workspace `display_name` naming standard, require a human approval step before change-control `start`. Publish under `examples/policies/`. | **P1** |
| **Run tasks** | A pre-apply run task that dry-runs the CVP workspace build (or checks build status) and blocks the run on a failed build — catches config errors before apply. | **P2** |
| **No-code modules** | Publish self-service CVP change modules (e.g., "apply a rack BGP change") to the private registry so operators fill a form instead of writing HCL. | **P2** |
| **Terraform Stacks** | Model multi-workspace CVP change management as a **Stack** — one component per CVP workspace — instead of workspace sprawl. Aligns with `design.md` D2 (one workspace = one unit) and the Terragrunt story (§4). | **P2** |
| **Private registry publishing** | Beyond the public Terraform/OpenTofu registries (`docs/release.md`), document private-registry distribution for on-prem/air-gapped consumers. | **P1** |

## 4. Terragrunt 1.1

| Capability | CVP fit | Prio |
|---|---|---|
| **Stacks (`terragrunt.stack.hcl`)** | One unit = one CVP workspace; a stack composes a change across units. See [`standards/terragrunt-integration.md`](standards/terragrunt-integration.md). | **P1** |
| **Redesigned catalog** | Reusable CVP change-unit templates (BGP bump, EVPN rollout) in a Terragrunt catalog for scaffolding. | **P2** |
| **Stack dependencies + read-based change detection** | Declare unit-to-unit wiring (inputs → change-control) up front; CI runs only units whose inputs changed. | **P1** |
| **Provider cache (OpenTofu 1.10+)** | Static `CGO_ENABLED=0` binary + correct `SHA256SUMS` so the Provider Cache Server works under `terragrunt run --all` (already a release requirement). | **P0** |

## 5. Reference implementations (learned from)

| Provider | What to borrow / avoid |
|---|---|
| [`terraform-provider-pgsteward`](../../terraform-provider-pgsteward/docs/research/tz-pgsteward.md) (spec) | **Borrow:** layered pure-core architecture, `TokenSource` auth, write-only + `*_wo_version`, ephemeral token resource, capability-matrix, object-identity-keyed serialization (the CVP analog of advisory locks — one workspace = one apply, `design.md` D2). |
| `cyrilgdn/terraform-provider-postgresql` | **Avoid:** SDKv2 → secrets in state (no write-only), inconsistent concurrency handling. Our whole reason to be greenfield on the framework. |
| Arista `terraform-provider-cloudeos` / CVaaS provider | **Reference:** closest domain analog for resource shapes; CVaaS-first is a documented non-goal (`design.md`), but a useful cross-check for surface parity. |
| `hashicorp/vault`, `Azure/azapi` | **Reference:** canonical ephemeral-resource + write-only implementations. |

## 6. Design spotlight: workflow verbs as Actions — ✅ RESOLVED (ADR 0006)

**Decision: B — resources + actions.** `cvp_workspace` holds declarative fields;
the workflow transitions are Terraform **Actions** (`ProviderWithActions`, 1.14)
the practitioner invokes via `lifecycle.action_trigger`. The v0.1
`auto_build` / `auto_submit` / `auto_approve` booleans are removed. This matches
CVP's actual imperative workflow and upholds the non-goal "do not automate
approval" (`design.md`).

The verb set is the `arista.workspace.v1` `Request` enum — build, cancel_build,
submit (+force), abandon, rollback, rebase. `approve` / `start` are **change
control** verbs, not workspace ones, and are tracked as a follow-up (§1).

Full rationale, live-lab evidence, and rejected alternatives:
[ADR 0006](adr/0006-workspace-resource-and-actions.md); design record
`design.md` D9.
