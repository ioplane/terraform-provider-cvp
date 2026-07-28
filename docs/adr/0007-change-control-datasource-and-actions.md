<div align="center">

# ADR 0007 — `cvp_change_control` as a data source + approve/start Actions

</div>

- **Status:** Accepted
- **Date:** 2026-07-29

## Context

A CVP **change control** (CC) groups the device tasks produced by a submitted
workspace into a schedulable, staged, RBAC-gated execution unit
(`arista.changecontrol.v1`). The v0.1 skeleton modelled it as a resource with
`auto_approve` / `wait_for_execution` booleans (design.md D3).

The generated stubs and a live lab probe reshape that:

- **Services.** `ChangeControlConfigService` (Set/GetOne/Delete) writes the CC
  config — `ChangeConfig{name, root_stage_id, stages, notes}`, a `start` flag,
  and a `schedule`. `ApproveConfigService` (Set/GetOne/Delete) writes the
  **approval**, pinned to a specific CC **version** (`ApproveConfig{approve,
  version, override_reason}`). `ChangeControlService` is **read-only**
  (GetOne/GetAll/Subscribe) — the status view (`status`, `error`, `device_ids`,
  validation, `completion_reason`).
- **Approval is a separate, version-pinned service** — the API itself enforces
  the separation of duties CVP documents as a "strict non-author review model"
  (arista-mcp: _Using Change Controls_). Approving requires the exact version
  you reviewed.
- **CCs are workspace-derived.** A change control is normally created by
  **submitting a workspace** (the `cvp_workspace_submit` action; the ids land in
  `cvp_workspace.cc_ids`), not authored field-by-field. Authoring a `ChangeConfig`
  (a nested `stages` map of actions and args) by hand is complex and rarely what
  a practitioner wants.
- **Evidence gate.** A live probe found **0 change controls** on the netlab2 lab
  (`ChangeControlService.GetAll` → empty). A CC only exists after a workspace
  with real device config changes is submitted, and _executing_ one pushes
  config to devices. So meaningful live acceptance needs a **safe CC fixture**
  (a workspace with a benign/no-op device change) — it is gated, not free.

## Decision

Model change control the same way ADR 0006 models the workspace verbs —
**declarative state read-only, imperative verbs as Actions** — rather than as a
mutating resource:

### Data source `cvp_change_control`

Read a CC's status by `id` (`ChangeControlService.GetOne`): computed `status`,
`error`, `device_ids`, `approved`, `started`, `completion_reason`. This is what
practitioners reference after a workspace submit (`cvp_workspace.cc_ids[0]`).

### Actions (`ProviderWithActions`, Terraform ≥ 1.14)

| Action | Service call |
|---|---|
| `cvp_change_control_approve` | `ApproveConfigService.Set{approve, version}` — version-pinned; `version` read from the CC state |
| `cvp_change_control_start` | `ChangeControlConfigService.Set{start}` — execute the (approved) CC |

`approve` is deliberately a separate action (never an `auto_approve` attribute),
upholding the separation-of-duties non-goal. A schedule/rollback action is
deferred until the rollback RPC surface and a safe fixture are confirmed.

### Deferred: authoring resource

A full `cvp_change_control` **resource** that authors a `ChangeConfig` from
scratch (opaque `change_json`, like `cvp_studio_inputs.inputs_json`) is deferred
to a later minor — it is the advanced case and needs the safe-fixture evidence
gate above.

## Consequences

- **Honest, consistent model.** Status is read (data source); approve/start are
  imperative Actions — symmetric with the workspace design (ADR 0006). No
  `auto_approve`; approval stays a deliberate, version-pinned, out-of-band step.
- **BREAKING vs v0.1.** The schema-only `cvp_change_control` _resource_ stub is
  replaced by a data source + actions. Pre-1.0 → MINOR bump with a `BREAKING:`
  changelog entry.
- **Implementation is evidence-gated.** Landing the data source + actions
  requires the safe CC fixture (submit a workspace with a no-op device change on
  the lab) for live acceptance; until then the client accessors + Action wiring
  are unit-tested and the live suite is added with the fixture.
- **`design.md` D3 is superseded** by this ADR (status stays computed; the
  `wait_for_execution` idea becomes a `check` block / poll on the data source,
  backlog §1).

## Alternatives rejected

- **Mutating `cvp_change_control` resource with `auto_approve` /
  `wait_for_execution` (v0.1):** overloads desired-state with an imperative,
  RBAC-gated approval; invites `auto_approve = true`. Rejected — same reasoning
  as ADR 0006.
- **Authoring the `ChangeConfig` as typed HCL:** the nested `stages`/actions tree
  is large and version-specific; typing it is brittle. If authored at all, it is
  an opaque `change_json` in a later minor.
