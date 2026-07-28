<div align="center">

# Testing

</div>

## Layers

| Layer | Build tag | Runner | Target |
|---|---|---|---|
| Unit | none | `go test -race ./...` | Schema validation, plan modifiers, CVP payload marshalling — no network. |
| Acceptance | `acceptance` + `TF_ACC=1` | `go test -tags acceptance ./...` | Real plan/apply/refresh/destroy against the netlab2 CVP lab via `terraform-plugin-testing`. |
| Matrix | external | CI matrix | Provider against multiple CVP releases (post-1.0). |
| Soak | external | continuous apply/refresh | Single-fabric stability (verification phase). |

## Unit tests

- `t.Parallel()` mandatory (`paralleltest`).
- Cover, per resource: empty config, max-length identifier, idempotent
  re-create, conflicting attributes, and a simulated gRPC error path.
- No live CVP — use in-process mocks / recorded payloads.

## Acceptance tests (`terraform-plugin-testing`)

- Gated by `TF_ACC=1`; **never** run in the default PR pipeline (they need lab
  credentials). Run locally or in a protected CI job with `gopass`-sourced
  secrets.
- **Not parallel** — a single shared live CVP. `paralleltest`/`tparallel` are
  excluded under `test/acceptance/`.
- Each resource's suite includes:
  - a `Create`→`Read` step asserting state via `statecheck` + `knownvalue`
    (not brittle string matches);
  - an **idempotency** check — `ConfigPlanChecks.PostApply` with
    `plancheck.ExpectEmptyPlan()`;
  - an **`ImportState`** step with `ImportStateVerify: true`
    (`ImportStateVerifyIgnore` only for write-only/secret attributes);
  - a `Destroy` with a `CheckDestroy` that asserts the CVP object is gone (no
    leaked workspace / abandoned change control).

```go
resource.Test(t, resource.TestCase{
    ProtoV6ProviderFactories: testAccProviderFactories,
    Steps: []resource.TestStep{
        {Config: cfg, ConfigPlanChecks: resource.ConfigPlanChecks{
            PostApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
        {ResourceName: "cvp_workspace.test", ImportState: true, ImportStateVerify: true},
    },
})
```

## Idempotence proof (per resource)

```text
1. terraform apply          → created
2. terraform apply          → empty plan
3. out-of-band CVP edit     → drift
4. terraform plan -refresh-only → drift reported
5. terraform apply          → remediated, empty plan after
```

## Lab safety

- Test object names are namespaced with the run id: `tf-acc-<RUN_ID>` — see
  `design.md` acceptance topology.
- `CheckDestroy` must leave zero residue; a leaked workspace fails the suite.
- Never point acceptance tests at a production CVP cluster.

## CI

- `task test` (unit + race) runs on every push/PR in `golang:1.26-trixie`.
- PR gate jobs (all SHA-pinned, see `.github/workflows/`): build-and-test,
  lint-go, lint-py (ruff + ty), vulncheck, **proto** (buf lint + `internal/pb`
  in-sync check), terraform (fmt), lint-docs, commitlint (PR title),
  **dependency-review**, **SonarQube** (coverage), plus the Security
  (CodeQL/gosec/Trivy/**Semgrep**) and Scorecard workflows.
- Acceptance tests are **out of the PR gate**; they run in a protected,
  credentialed job or locally via `task verify`.
- Coverage uploaded to Codecov; JUnit report via `gotestsum`.
