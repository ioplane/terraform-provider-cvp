<div align="center">

# Design — `provider::cvp::studio_path` provider-defined function

</div>

- **Status:** Approved (brainstorming)
- **Date:** 2026-07-29

## Problem

`cvp_studio_inputs.path` is a `list(string)` of bracket-notation segments into a
Studio input schema. Real studios mix three addressing modes that are easy to get
wrong by hand:

- **group members** — plain segments (`"vrfDescription"`, `"enabled"`);
- **keyed collections** — `"[<key>=<value>]"` (`"[tenantName=RED]"`,
  `"[vrfName=RED-VRF]"`, `"[vlanId=100]"`, `"[profileName=access]"`);
- **resolver segments** — a resolved device/interface/tag id.

Hand-writing the `[key=value]` brackets (and remembering the per-collection key
name) is error-prone. A pure helper that assembles the path array removes that
class of mistake.

## Non-goals / constraints

- **Generic only.** No hardcoded studio ids, collection-key names, or any
  Uzum-specific data. The function knows nothing about any particular studio.
- **No schema validation.** The function does not verify a path against a real
  studio schema — CVP validates on write (design.md D1/D5). The function only
  assembles/normalizes the array.
- **Pure & offline.** No CVP calls, no I/O; deterministic. (HashiCorp design
  principles 6/7.)
- **Later customization stays additive** (a wrapping Terraform module with a
  studio's key names; or future optional args) — the generic core does not
  change.

## Signature

```hcl
provider::cvp::studio_path(segment...) -> list(string)
```

- `segment` — variadic, **dynamic**. Each argument is either:
  - a **string** → emitted verbatim (group member / resolver id):
    `"vrfDescription"` → `"vrfDescription"`;
  - a single-key **object** → emitted as a bracket key segment:
    `{ vrfName = "RED-VRF" }` → `"[vrfName=RED-VRF]"`;
    `{ vlanId = 100 }` → `"[vlanId=100]"` (number stringified).
- Zero arguments → `[]` (the studio root).

Example:

```hcl
path = provider::cvp::studio_path(
  "tenants", { tenantName = "RED" },
  "vrfs",    { vrfName = "RED-VRF" },
  "vrfDescription",
)
# => ["tenants","[tenantName=RED]","vrfs","[vrfName=RED-VRF]","vrfDescription"]
```

## Implementation

- `internal/functions/studio_path.go`
  - `type studioPathFunction struct{}` implementing `function.Function`.
  - `Metadata` → name `studio_path`.
  - `Definition` → `VariadicParameter: function.DynamicParameter{Name: "segment", ...}`,
    `Return: function.ListReturn{ElementType: types.StringType}`.
  - `Run`:
    1. read the variadic tuple into `[]types.Dynamic`;
    2. for each, inspect the underlying value:
       - `types.String` → the segment string;
       - `types.Object` → require exactly one attribute; format
         `"[" + key + "=" + stringify(value) + "]"`; zero/multiple keys →
         `function.NewArgumentFuncError(pos, ...)`;
       - `types.Number` → allowed only inside an object value; a bare number/
         bool/other segment → argument error (segments are strings or objects);
       - null/unknown underlying → argument error (paths must be known);
    3. `stringify`: `types.String` → value; `types.Number` → integer/plain
       decimal; `types.Bool` → `"true"`/`"false"`; else argument error.
    4. set the result list.
- Wire `provider.ProviderWithFunctions.Functions(ctx) []func() function.Function`
  in `internal/provider/provider.go`.
- Compile-time assertions: `var _ function.Function`, `var _
  provider.ProviderWithFunctions`.

### Units & boundaries

The function is a single, self-contained unit: input = variadic dynamic
segments, output = `list(string)`, no dependencies on the CVP client. The
segment-formatting logic (`segmentToString`, `stringify`) is factored as pure
helpers so it is unit-tested without the framework harness where possible.

## Testing

- **Unit** (`internal/functions/studio_path_test.go`): table over the pure
  helpers — group member, string-keyed collection, int-keyed collection, bool
  value, root (empty), and the error cases (multi-key object, empty object, bare
  number segment). Uses the framework's function test facilities for the `Run`
  path.
- **`terraform test`** (`.tftest.hcl`) or an `examples/functions/studio_path/`
  example exercising the function end-to-end (fmt-checked, tfplugindocs-rendered).
- Requires **Terraform ≥ 1.8** for provider-defined functions (documented in the
  example's `required_version`).

## Docs & plan

- Regenerate registry docs (`task docs` renders `docs/functions/studio_path.md`).
- `docs/backlog.md` §1: mark the provider-defined-function row done for
  `studio_path`; keep `cvp_studio` (generic studio-definition resource) and D6
  write-only secret inputs as generic follow-ups (no Uzum coupling).
- `README.md`: add a short "Functions" section; `CHANGELOG.md` Unreleased entry.

## Verification

All in-container via Task (project rule): `task all` (build, test, golangci-lint,
lint-docs incl. regenerated docs, vulncheck) green; `task tffmt-check`. No live
lab needed (pure function).
