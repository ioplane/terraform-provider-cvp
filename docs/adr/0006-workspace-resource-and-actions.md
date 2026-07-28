<div align="center">

# ADR 0006 — `cvp_workspace` as a declarative resource + workflow Actions

</div>

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

A CVP workspace has two distinct kinds of behaviour that the v0.1 skeleton
conflated into one resource:

1. **Declarative desired state** — a workspace has a `display_name`, a
   `description`, an `exclude_network_provisioning` flag, and a stable
   server-visible `workspace_id`. This is a good fit for a Terraform _resource_:
   `terraform apply` converges CVP to the described config.
2. **Imperative workflow verbs** — _build_, _submit_, _abandon_, _rollback_,
   _rebase_, _cancel build_. These are one-shot operations against the CVP
   change-management state machine. They are **not** desired state: "build" is
   not a property a workspace _has_, it is a thing you _do_ to it, once.

The v0.1 skeleton modelled the verbs as boolean attributes (`auto_build`,
`auto_submit`, `auto_approve`) on the resource. That overloads a declarative
resource with side effects hidden inside state transitions, makes the operations
non-repeatable (you cannot "build again" without a config churn), and quietly
encourages `auto_approve = true`, which violates the separation-of-duties
non-goal in [`design.md`](../../design.md). Backlog §6 flagged this as a
contract-affecting decision to resolve at the design gate.

Terraform **Actions** (`provider.ProviderWithActions`, Terraform ≥ 1.14,
`terraform-plugin-framework` ≥ v1.15) exist precisely for ad-hoc, imperative
side effects invoked from configuration. They are the honest model for the CVP
verbs.

### Evidence (live netlab2 CVP lab)

The design is grounded in a live round-trip against the lab, not the proto
shapes alone (methodology "evidence discipline"):

| Operation | Call | Observed result |
|---|---|---|
| create | `WorkspaceConfigService.Set{key, display_name, description}` | `state = WORKSPACE_STATE_PENDING`, `needs_build = false` |
| read | `WorkspaceService.GetOne{key}` | full state (`state`, `needs_build`, `last_build_id`, `cc_ids`, audit timestamps) |
| build | `Set{request = REQUEST_START_BUILD, request_params.request_id}` | async build; the supplied `request_id` becomes `last_build_id` |
| abandon | `Set{request = REQUEST_ABANDON, request_params.request_id}` | `state → WORKSPACE_STATE_ABANDONED` |
| destroy | **abandon**, then `WorkspaceConfigService.Delete{key}` | subsequent `GetOne → NotFound` (fully removed) |

Two findings shaped the contract:

- **Every `Request` verb requires `request_params.request_id`.** `Set` with a
  `request` but no `request_id` returns `InvalidArgument: request ID parameter
  is missing`. The provider therefore mints a UUIDv4 `request_id` for every
  workflow invocation.
- **Config `Delete` alone does not tear a workspace down.** On a `PENDING`
  workspace, `Delete` leaves `GetOne` still returning `PENDING`. Only
  **abandon → Delete** yields `NotFound`. So resource `Delete` abandons a
  non-terminal workspace first, then deletes the config.

### Verb set is API-derived (correcting the requirement)

The `arista.workspace.v1` `Request` enum is the ground truth for workspace
verbs:

```text
REQUEST_START_BUILD  REQUEST_CANCEL_BUILD  REQUEST_SUBMIT
REQUEST_SUBMIT_FORCE REQUEST_ABANDON       REQUEST_ROLLBACK  REQUEST_REBASE
```

There is **no `approve` or `start` verb on a workspace** — those belong to
_change control_ (`arista.changecontrol.v1`), a separate resource. The informal
"build / submit / approve / start / abandon" wording mixed the two state
machines; this ADR scopes the workspace Actions to the workspace enum above and
defers `approve`/`start` to the change-control work.

## Decision

Model `cvp_workspace` as a **declarative resource** and the workflow verbs as
**provider Actions**.

### Resource `cvp_workspace`

- **Arguments:** `workspace_id` (optional + computed — client-minted UUIDv4 when
  omitted; `RequiresReplace`), `display_name` (required), `description`
  (optional), `exclude_network_provisioning` (optional + computed).
- **Computed / read-only:** `state`, `needs_build`, `last_build_id`,
  `created_at`, `created_by`, `last_modified_at`, `last_modified_by`, `cc_ids`.
- **CRUD:** Create/Update → `WorkspaceConfigService.Set`; Read →
  `WorkspaceService.GetOne` (maps `NotFound` to state removal); Delete →
  abandon-then-`Delete`.
- **Import:** by `workspace_id`.
- The `auto_build` / `auto_submit` / `auto_approve` booleans are **removed**
  (BREAKING; pre-1.0, MINOR bump).

### Actions (`ProviderWithActions`)

Unlinked actions, one gRPC `Set{request, request_id}` each, sharing one client
helper:

| Action | `Request` | Extra config |
|---|---|---|
| `cvp_workspace_build` | `REQUEST_START_BUILD` | — |
| `cvp_workspace_cancel_build` | `REQUEST_CANCEL_BUILD` | — |
| `cvp_workspace_submit` | `REQUEST_SUBMIT` / `REQUEST_SUBMIT_FORCE` | `force` (bool) |
| `cvp_workspace_abandon` | `REQUEST_ABANDON` | — |
| `cvp_workspace_rollback` | `REQUEST_ROLLBACK` | — |
| `cvp_workspace_rebase` | `REQUEST_REBASE` | — |

Each action takes `workspace_id` (required) and mints its own `request_id`,
surfaced via a progress event. No `approve` action — that is change control's.

## Consequences

- **Honest model:** declarative config and one-shot verbs are separated the way
  CVP separates them; verbs are repeatable and free of config churn.
- **Separation of duties preserved:** there is no `auto_approve`; approval stays
  a deliberate, out-of-band step (design.md non-goal upheld). Policy-as-code can
  gate the Actions (backlog §3).
- **Minimum Terraform bumps to 1.14** for consumers who use the Actions; the
  resource alone still works on older core. Documented in the compatibility
  matrix; acceptance for Actions runs on the container's Terraform 1.15.8
  (OpenTofu 1.12 lacks Actions and is skipped for those tests).
- **BREAKING schema change** vs v0.1 (booleans removed). Pre-1.0, shipped as a
  MINOR bump with a `BREAKING:` changelog entry (versioning standard).
- **Cost:** six near-identical action types. Mitigated by a shared
  `internal/client/cvp` request helper and a common invoke path; the distinct
  type names are kept for discoverability and per-verb schemas.

## Alternatives rejected

- **Resources-only (v0.1 booleans):** hides imperative side effects in state,
  non-repeatable, invites `auto_approve`. Rejected — the reason for this ADR.
- **A single `cvp_workspace_request` action with a `request` string
  attribute:** fewer types, but untyped, less discoverable, and cannot carry
  per-verb schema (e.g. `submit.force`). Rejected in favour of typed actions.
- **A `cvp_workspace_build` _resource_** (build-as-state): a build is a
  point-in-time event, not durable state; re-running it via config edits is
  awkward. Rejected in favour of an action.
