# terraform-provider-cvp — design

**Дата**: 2026-07-16
**Scope**: I5 phase из `docs/plans/2026-07-16-cvp-integration-plan.md`.
**Baseline**: existing `terraform-provider-cvaas` (Arista official) + `arista-cloudvision`
провайдер сообщества (deprecated).

## Не-цели

- **Не заменяем CVaaS provider**. Тот — SaaS-first, CVaaS-specific resource shapes.
- **Не автоматизируем approval**. Workflow ports от CVP на 1:1 в HCL с `auto_approve = false`
  default. Human approval через UI или CI approval step.
- **Не поддерживаем CVaaS resources без on-prem equivalent** (asset_manager tenant assign etc.).
  Feature-flag: `provider "cvp" { cvaas = true|false }` — false default.

## Ключевые design decisions

### D1 — Path-overlap detection на plan time

**Rule**: два `cvp_studio_inputs` не могут share prefix. Overlap = destructive Set order-dependency.

**Implementation** (v0.2):

```go
type ConfigValidator struct{}
func (v ConfigValidator) ValidateResource(ctx, req, resp) {
    // Aggregate all cvp_studio_inputs.path values across config
    // Sort lexicographically
    // For each adjacent pair, error if one is prefix of другой
}
```

Требует `provider.ConfigValidator` interface. Framework поддерживает.

### D2 — Workspace = one apply

`cvp_workspace` — root of dependency chain. Все Studio inputs, tags, CC resources depend on него
(`workspace_id` reference).

**Не supported**: multiple workspaces в one apply. Terragrunt module = one workspace.

Rationale: workspace conflict semantic (I2 finding — path conflict resolution to last-approved wins)
плохо interacts с parallel resources.

### D3 — CC status как computed attribute

`cvp_change_control.status` — computed only. Options:
- `PENDING` — created, awaiting approval
- `APPROVED` — approved, awaiting start
- `RUNNING` — executing
- `SUCCESS` / `ROLLED_BACK` / `FAILED` — terminal

If `wait_for_execution = true` — provider blocks until terminal state.
Timeout: `changecontrol_timeout` provider argument, default 30m.

### D4 — Три auth methods, native provider config

`auth_method = "session" | "cert" | "bearer"`.

**session**: token pre-obtained (CI/CD login step), passed via `token`.
**cert**: `cert_pem`, `key_pem`, optional `ca_pem`.
**bearer**: `token` — long-lived API token.

Rejected alternatives:
- Provider-side session login (username+password). **Rejected** — password в state, credentials risk.
- Environment-only config. **Rejected** — poor UX with Terragrunt.

### D5 — Immutable studios (from_package)

`cvp_studio.from_package` — computed. Если non-empty после import, provider marks resource как
read-only + errors on Update. Rationale — CVP rejects modification anyway; fail early с clear
error лучше than gRPC error at apply time.

### D6 — Secrets

`is_secret: true` fields в Studio input schema — provider **не read'ит** unmasked value.
`SecretInputService.GetOne` remain accessible только через CVP UI + specialized ops tooling.

Value passing:
- Write: user provides plaintext в `inputs_json` (Terraform state contains plaintext).
- Read: provider skips SecretInputs during Read → drift detection disabled для secret fields.
- Alternative для production: reference `data.hashicorp_vault_kv_secret` в `inputs_json` composition.

### D7 — Retry policy

- Transient (`Unavailable`, `DeadlineExceeded`): exponential backoff, 5 attempts, 1-16s base.
- Semantic (`AlreadyExists`, `NotFound`, `FailedPrecondition`): fail immediately.
- Auth (`Unauthenticated`, `PermissionDenied`): fail immediately + suggest auth method check.

Implementation: `google.golang.org/grpc/retry` middleware.

### D8 — Prototype vs production gates

**v0.1 (prototype, this repo state)**:
- 3 resources, schema authored, CRUD stubbed.
- No gRPC calls yet.
- Not usable для actual TF apply.

**v0.2 (usable prototype)**:
- Full CRUD wired.
- One acceptance test per resource (against um-cvp lab).
- Documented breaking changes probability.

**v0.3 (production-track)**:
- Retry + timeout knobs.
- All 10 Tier 1 resources.
- Full acceptance test suite.
- SemVer commitment + CHANGELOG.

**v1.0 (production)**:
- 3-tier resources.
- Multi-cluster support (parametrize provider by cluster).
- SBOM + signed releases via cosign.

## Acceptance test topology

```
um-cvp01 (172.29.0.54)  ←── existing lab кластер
  │
  ├─ Studio EVPN-ESI (UUID stable)
  ├─ Test workspace: tf-acc-{RUN_ID}
  └─ Test rack tag: tf-acc-rack-{RUN_ID}
```

Test lifecycle per resource:
1. Create workspace.
2. Set inputs at test path.
3. Assert Read returns матching inputs.
4. Delete workspace via abandon.
5. Assert no leaked resources.

## Cross-project glue (I6 preview)

- `um-mgt/live/s1/mgt/cvp/main.tf` — first consumer.
- Depends on `pulumi_stack.cvp.outputs.cluster_endpoint` (см. I3 §9).
- Migration path через `terraform import` + `migrations/ownership.yaml`.

## Open questions

1. **Batch RPCs** — `SetSome`, `SetAll` supported for InputsConfigService? Uncertain from public proto —
   default только `Set` on per-key basis. Batch would drastically reduce RPC count per apply.
2. **Subscription vs polling** — build/CC status. `Subscribe` more efficient, but harder to reconcile
   с Terraform's sequential apply model. **Decision**: polling для v0.2, evaluate Subscribe для v0.3.
3. **Provider identity** — registry publishing. `ioplane/cvp` is placeholder; needs registration
   в HashiCorp Registry OR private registry (`private.terraform.io/ioplane/cvp`).

## Follow-on tasks

- **I5-grpc-wire**: implement CRUD hooks (P1 in code).
- **I5-acc-test**: acceptance test harness (docker-compose с um-cvp01 kubectl-forwarded).
- **I5-tf-registry**: publish v0.1 to private registry.
- **I5-cvaas-parity**: cross-check вertaform-provider-cvaas для surface gaps.
