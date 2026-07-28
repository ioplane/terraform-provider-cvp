<div align="center">

# Changelog

</div>

All notable changes to `terraform-provider-cvp` are recorded in this file.

The format follows [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).
Commit messages follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).

While the provider is pre-`1.0.0`, the schema and public surface may change
between minor versions; breaking changes bump the MINOR component per SemVer §4
and are called out under a `Changed` heading with a `BREAKING:` prefix.

## [Unreleased]

### Added

- **Engineering standards and repository governance.** Ported the house
  standard from the sibling `pulumi-eos` and adapted it for a Terraform
  provider:
  - Podman/OCI dev container (`golang:1.26-trixie`) with the full toolchain
    (Go 1.26.5, golangci-lint v2.12.2, govulncheck, osv-scanner, Terraform
    1.15.8, OpenTofu 1.12.5, Terragrunt 1.1.1, tfplugindocs, goreleaser, buf,
    uv/ruff/ty), driven by `podman-compose` (`deployments/`).
  - `golangci-lint` v2 allowlist (~90 linters, severity-tiered) tuned for the
    Terraform Plugin Framework.
  - `Taskfile.yml` (container-driven), `.goreleaser.yml` v2 (GPG-signed
    `SHA256SUMS` + SPDX SBOM per the Terraform Registry contract), and
    SHA-pinned GitHub Actions workflows (CI, release, security, Scorecard,
    dependency-review) plus Dependabot; SAST via **Semgrep** — `task semgrep`
    (host, Pro rules via login) for local runs, and a `semgrep ci` CI job using
    the `SEMGREP_APP_TOKEN` secret (Pro rules); **SonarCloud** static analysis
    (`sonar-project.properties` + a SonarQube CI job with Go coverage, using the
    `SONAR_TOKEN` secret); `.env.example` documents the local token variables.
  - Governance: `AGENTS.md` (+ `CODEX.md` / `CLAUDE.md` pointers), Apache-2.0
    `LICENSE`, `CONTRIBUTING.md`, `GOVERNANCE.md`, `SECURITY.md`, `SUPPORT.md`,
    `CODE_OF_CONDUCT.md`, `CODEOWNERS`, issue/PR templates, `VERSION`, and this
    changelog.
  - Standards docs under `docs/standards/` (naming, versioning, commits,
    changelog, Go 1.26 style, Terraform-provider best practices,
    Terragrunt integration) and a chosen delivery methodology
    (`docs/methodology.md`).
  - Mandatory git-worktree + Pull-Request-only workflow (`scripts/worktree.sh`);
    branch protection ruleset on `main`.
- **CVP gRPC stubs (`internal/pb`).** Vendored the `arista.workspace.v1` proto
  import closure from `cloudvision-apis` (pinned revision) under `api/proto/` and
  generated the Go message + gRPC-client stubs via `buf` (`task proto`,
  deterministic). First step of the P1 gRPC wiring (ADR 0005); `WorkspaceConfigService.Set/Delete`
  and `WorkspaceService.GetOne` are now available to `internal/client/cvp`.
- **`arista.studio.v1` gRPC stubs.** Vendored `arista/studio.v1/{studio,services.gen}.proto`
  at the same pinned revision (no new transitive imports) and generated the Go
  stubs. The upcoming `cvp_studio_inputs` resource is backed by
  **`InputsConfigService`** (GetOne/GetAll/Set/Delete); secret inputs are written
  through it as **write-only** attributes (design.md D1/D5/D6). The generated
  read-only `SecretInputService` (Get/Subscribe only, returns the _unmasked_
  value) is **deliberately not consumed** by the provider — reading it would pull
  plaintext secrets into provider memory and violate D6; it is reserved for UI /
  ops tooling.
- **`internal/client/cvp` connection foundation.** A gRPC/TLS client with the
  three auth methods (design.md D4, evidence: the CVP authz model in
  `arista-cvp-re`): `bearer`/`session` attach `authorization: Bearer <token>`
  per-RPC, `cert` uses client mTLS; TLS 1.3 minimum, optional pinned CA. Transient
  failures (`UNAVAILABLE`) retry with exponential backoff (D7). The provider
  `Configure` now builds the client and surfaces config errors as attribute
  diagnostics. Unit-tested (config validation, bearer metadata, TLS creds).
- **`cvp_workspace` — full CRUD, import, and workflow Actions (ADR 0006).** The
  workspace is now a real, usable resource: Create/Update via
  `WorkspaceConfigService.Set`, Read via `WorkspaceService.GetOne` (with drift
  and `NotFound` handling), Delete via evidence-backed **abandon-then-config-delete**,
  and `ImportState` by `workspace_id`. New computed attributes surface server
  state (`state`, `needs_build`, `last_build_id`, `created_at`/`created_by`,
  `last_modified_*`, `cc_ids`). The imperative workflow verbs are **Terraform
  Actions** (`ProviderWithActions`, Terraform ≥ 1.14): `cvp_workspace_build`,
  `cvp_workspace_cancel_build`, `cvp_workspace_submit` (with `force`),
  `cvp_workspace_abandon`, `cvp_workspace_rollback`, `cvp_workspace_rebase` —
  each one `Set` with a `Request` enum plus a minted `request_id`. Unit-tested
  (state mapping, verb→request table, request-id format) and covered by **live
  acceptance** against the netlab2 CVP lab (create/read/update/import/destroy +
  build/abandon verb lifecycle). Recorded in [ADR 0006](docs/adr/0006-workspace-resource-and-actions.md)
  and `design.md` D9.
- **`cvp_studio_inputs` — full CRUD and import.** A declarative inputs value at
  a Studio path inside a workspace, backed by `arista.studio.v1.InputsConfigService`:
  Create/Update via `Set`, Read via the config `GetOne` (which round-trips the
  written value exactly), Delete via the config `Delete`, and `ImportState` by
  the composite id `{workspace_id}/{studio_id}/{path...}`. The key
  (`studio_id`, `workspace_id`, `path`) is immutable (`RequiresReplace`);
  `inputs_json` is validated as JSON at plan time, and CVP validates it against
  the studio schema on write. Prefix-overlapping sibling paths (which CVP would
  silently clobber, a higher `Set` overwriting lower entries) are rejected before
  a write (design.md D1); the import id percent-escapes each segment so resolver
  paths containing `/` (e.g. `[tags/query=…]`) round-trip. Live acceptance
  against the netlab2 lab (create/read/idempotency/update/import/destroy at a
  studio root). Secret inputs (write-only, design.md D6) are a tracked follow-up
  (backlog §1).
- **`provider::cvp::studio_path` provider-defined function (Terraform ≥ 1.8).**
  Assembles a `cvp_studio_inputs` `path` (`list(string)`) from ordered variadic
  segments: a string is emitted verbatim (group member / resolver id), a
  single-key object becomes bracket key-notation (`{ vrfName = "RED-VRF" }` →
  `"[vrfName=RED-VRF]"`, numbers stringified so `{ vlanId = 100 }` →
  `"[vlanId=100]"`); no segments yields `[]` (the studio root). Generic, pure and
  offline — it does not validate against any studio schema (CVP does that on
  write). Unit-tested and exercised end-to-end via `terraform apply`.
- **Compatibility contract (`docs/compatibility.md`).** A consumer-facing matrix
  of the minimum Terraform / OpenTofu core version per capability — base
  resources, provider-defined functions (Terraform ≥ 1.8 / OpenTofu ≥ 1.7), and
  **Actions (Terraform ≥ 1.14; not available in OpenTofu)** — plus the wire
  protocol (6.0), framework and Go versions. Linked from the README and docs
  index; `commitlint` now also allows the `functions` and `actions` commit
  scopes.
- **`arista.changecontrol.v1` gRPC stubs.** Vendored
  `arista/changecontrol.v1/{changecontrol,services.gen}.proto` at the pinned
  revision (no new transitive imports) and generated the Go stubs. The writable
  config services `ChangeControlConfigService` and `ApproveConfigService`
  (GetOne/Set/Delete) back the upcoming `cvp_change_control` resource and its
  approve/start/rollback actions; the read-only `ChangeControlService`
  (GetOne/GetAll/Subscribe) is the state view for computed status (design.md D3,
  ADR 0006).
- **Change-control design frozen ([ADR 0007](docs/adr/0007-change-control-datasource-and-actions.md)).**
  `cvp_change_control` will be a **data source** (computed `status`, `error`,
  `device_ids`) plus **approve / start Actions** (`ApproveConfigService.Set`
  version-pinned; `ChangeControlConfigService.Set` start flag) — not a mutating
  resource with `auto_approve` / `wait_for_execution`, upholding the
  separation-of-duties non-goal (symmetric with ADR 0006). Supersedes the v0.1
  `cvp_change_control` resource stub and `design.md` D3. Implementation is
  evidence-gated: a live probe found **0 change controls** on the lab (they are
  created by submitting a workspace with real device changes), so live
  acceptance needs a safe fixture first.
- **Capability backlog** — `docs/backlog.md` maps modern Terraform (Actions,
  managed identity, ephemeral/write-only, functions, `terraform test`),
  Terragrunt 1.1 (stacks/catalog), and HCP/TFE enterprise (dynamic OIDC
  credentials, Sentinel/OPA, run tasks, no-code modules, Stacks) capabilities to
  CVP, with priorities. Flags the "workflow verbs as Actions" contract question.

### Changed

- **BREAKING: `cvp_workspace` schema reworked (ADR 0006).** The `auto_build`,
  `auto_submit` and `auto_approve` booleans are **removed** — the workflow verbs
  are Terraform Actions now (see Added). The stable key attribute is renamed from
  `id` to **`workspace_id`** (optional; provider-minted when omitted). Import is
  by `workspace_id`. Pre-1.0, shipped as a MINOR bump per the versioning
  standard.
- **Build runner migrated from `make` to [Task](https://taskfile.dev)**
  (`Taskfile.yml` replaces the `Makefile`); all docs and hooks updated.
- **`scripts/automation/build.py` now uses the `podman-py` library** (Podman
  REST socket) for the image build + container lifecycle instead of shelling
  out, and is linted with **ruff** + type-checked with **ty**, run via **uv**
  (`task lint-py` / `task fmt-py`; `pyproject.toml` added; CI job + pre-commit
  hook added).
- **Dependencies bumped to latest (July 2026):** Go directive `1.26.5`;
  `terraform-plugin-framework` `v1.13.0` → `v1.19.0`; `grpc` `v1.78.0` →
  `v1.82.1` (clears the GO-2026-4762 advisory noted in the v0.1 skeleton);
  `protobuf` `v1.36.11`; transitive `golang.org/x/{net,text}` bumped to patched
  releases (govulncheck: 0).
- **Documentation presentation:** every Markdown file's top title is centered,
  and the README badges use [shieldcn](https://shieldcn.dev) shadcn/ui-styled
  badges (`variant=secondary`) centered in the header.

### Removed

- **`aristanetworks/cloudvision-go` dependency dropped for good.** It dragged a
  conflicting ancient `google.golang.org/genproto` (broke `go mod tidy`) and
  pulled the entire CloudVision Go surface for the sake of three services.
  Replaced by a **custom buf-generated gRPC client** (ADR 0005): vendored protos
  under `api/proto/`, generated stubs in `internal/pb/` via `buf`
  (`buf.yaml` + `buf.gen.yaml`, pinned `protoc-gen-go` v1.36.11 /
  `protoc-gen-go-grpc` v1.6.2), `task proto` / `proto-lint` / `proto-breaking`,
  and `buf` added to the dev image. Actual stub generation is the first P1
  gRPC-wiring step; the provider builds dependency-free until then.

### Fixed

- **`task testacc` now passes `-tags acceptance`** and forwards
  `CVP_ENDPOINT` / `CVP_AUTH_METHOD` / `CVP_TOKEN` through the container exec, so
  the acceptance suite actually compiles/runs and authenticates (review finding).
- **Security workflow jobs retain `contents: read`** — a job-level `permissions`
  mapping replaces the workflow-level one, so CodeQL/gosec/Trivy needed it added
  explicitly (review finding).
- **CI container build step pinned to `bash`** (dash lacked the `${VAR::N}`
  substring expansion); dependency-review switched to a copyleft deny-list to
  avoid a false positive on Google's compound-licensed Go modules.
- Clarified that the golangci-lint gate is zero-findings (any finding fails the
  run); the severity tiers classify findings, not the exit code (review finding).
- Pinned `MinVersion: tls.VersionTLS13` on the `insecure_tls` credentials path
  too (Semgrep `missing-ssl-minversion`).

## [0.1.0] — 2026-07-16

### Added

- Initial `terraform-provider-cvp` v0.1 skeleton: provider server bootstrap,
  provider schema (`endpoint`, `auth_method`, `cert`/`session`/`bearer`), and
  three schema-only resources with stubbed CRUD (`cvp_workspace`,
  `cvp_studio_inputs`, `cvp_change_control`). See `design.md`.

[Unreleased]: https://github.com/ioplane/terraform-provider-cvp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ioplane/terraform-provider-cvp/releases/tag/v0.1.0
