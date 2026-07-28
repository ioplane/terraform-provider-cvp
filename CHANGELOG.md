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

### Changed

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
  and the README badges are reworked to shields.io flat-style badges centered in
  the header.

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
