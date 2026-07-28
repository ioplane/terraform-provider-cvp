<div align="center">

# Terraform provider best practices

</div>

The rules the provider is held to, distilled from HashiCorp's
[provider design principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles),
the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
docs, and this project's target (Arista CloudVision Portal over gRPC/TLS).
Grounded against the framework via the `context7` MCP.

## 1. Design principles (HashiCorp)

1. **One problem domain.** This provider manages CloudVision Portal (on-prem);
   CVaaS-only shapes are out of scope (a documented non-goal, see `design.md`).
2. **One API object per resource.** `cvp_workspace`, `cvp_studio_inputs`,
   `cvp_change_control` each map to one CVP service object. Orchestration across
   objects belongs in Terraform modules / Terragrunt, not in a mega-resource.
3. **Schema mirrors the API.** Attribute names track CVP field names where they
   don't hurt UX; simplifications live in modules, not provider magic.
4. **Every resource is importable.** `ResourceWithImportState` is mandatory —
   brownfield CVP fabrics must be adoptable via `terraform import`.
5. **State continuity + SemVer.** State shape is a contract; changes ship a
   state upgrader. Versioning per [`versioning.md`](versioning.md).
6. **Provider functions are pure and offline.** Anything needing the network is
   a data source, not a function.
7. **Ephemeral resources for sensitive, non-persisted data** (e.g. a
   short-lived API token minted for a run) — never write secrets to state that
   don't need to be there.

## 2. Framework conventions

- Use **`terraform-plugin-framework`** (v1.19+), protocol 6. Not the legacy SDKv2.
- Implement the optional interfaces deliberately:
  - `ResourceWithConfigure` — receive the `*ClientBundle` from the provider.
  - `ResourceWithImportState` — always.
  - `ResourceWithModifyPlan` — for computed-on-change and `RequiresReplace`
    logic (e.g. immutable `from_package` studios, see `design.md` D5).
  - `ResourceWithConfigValidators` / `ValidateConfig` — cross-attribute rules
    (e.g. `auth_method=cert` requires `cert_pem`+`key_pem`; studio-inputs path
    overlap detection, `design.md` D1).
  - `ResourceWithUpgradeState` — whenever stored state shape changes.
- Prefer **declarative validators** from
  `terraform-plugin-framework-validators` over hand-rolled checks.
- Mark secrets `Sensitive: true`; use **write-only** attributes (Terraform 1.11+)
  for values that must never persist to state where drift-detection isn't needed.
- Every attribute has a `MarkdownDescription`. Docs are generated from the
  schema by `tfplugindocs` — the schema is the source of truth.
- Return a `DetailedDiff`/plan that honours `DryRun`; a re-apply of the same
  config must produce an empty plan (idempotency).

## 3. Error, retry, and logging behaviour

- **Diagnostics, not panics.** Surface actionable messages via
  `resp.Diagnostics.AddError` / `AddAttributeError`.
- **Retry policy** (`design.md` D7): transient gRPC codes (`Unavailable`,
  `DeadlineExceeded`) → exponential backoff (5 attempts, 1–16 s); semantic codes
  (`AlreadyExists`, `NotFound`, `FailedPrecondition`) → fail fast; auth codes
  (`Unauthenticated`, `PermissionDenied`) → fail fast with an auth-method hint.
- **Logging** via `tflog` with structured key-value fields; never log secrets.

## 4. Testing

- **Acceptance tests** (`terraform-plugin-testing`, `TF_ACC=1`) exercise real
  plan/apply/refresh/destroy against the netlab2 CVP lab. Every resource has at
  least one.
- Use **plan checks** (`plancheck.ExpectEmptyPlan` post-apply for idempotency)
  and **state checks** (`statecheck` + `knownvalue`) rather than brittle string
  asserts.
- Always include an **`ImportState` step** with `ImportStateVerify: true`.
- Unit tests cover schema validation, plan modifiers, and CVP-payload
  marshalling without a network.
- Details in [`../testing.md`](../testing.md).

## 5. Definition of done (per resource)

1. `Schema` with `MarkdownDescription` on every attribute; secrets `Sensitive`.
2. `Create` / `Read` / `Update` / `Delete` implemented against the CVP client.
3. `ImportState` implemented.
4. Plan is idempotent (same config twice → empty plan) and honours `DryRun`.
5. `ModifyPlan` / validators for cross-field rules and immutability.
6. `UpgradeState` if the state shape changed.
7. ≥ 5 unit edge cases (empty, max-length id, idempotent re-create, conflict,
   network error).
8. ≥ 1 acceptance test incl. an `ImportState` verify step, green against the lab.
9. Drift: an out-of-band CVP edit is reported by `terraform plan -refresh-only`.
10. Registry docs regenerated (`task docs`) and `task docs-check` clean.
11. `CHANGELOG.md` `[Unreleased]` updated.
12. New Arista-behaviour claims corroborated via `arista-mcp` and cited.

## 6. Documentation & registry

- `tfplugindocs` generates `docs/` from schema + `examples/` + `templates/`.
- Registry publishing needs: Apache-2.0 license (present), a valid top-level
  `terraform-registry-manifest.json` (protocol 6.0), a GPG-signed release, and
  example programs. See [`../release.md`](../release.md).
