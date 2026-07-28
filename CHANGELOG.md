# Changelog

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
    1.15.8, OpenTofu 1.12.5, Terragrunt 1.1.1, tfplugindocs, goreleaser),
    driven by `podman-compose` (`deployments/`).
  - `golangci-lint` v2 allowlist (~90 linters, severity-tiered) tuned for the
    Terraform Plugin Framework.
  - `Makefile` (container-driven), `.goreleaser.yml` v2 (GPG-signed
    `SHA256SUMS` + SPDX SBOM per the Terraform Registry contract), and four
    SHA-pinned GitHub Actions workflows (CI, release, security, Scorecard) plus
    Dependabot.
  - Governance: `AGENTS.md` (+ `CODEX.md` / `CLAUDE.md` pointers), Apache-2.0
    `LICENSE`, `CONTRIBUTING.md`, `SECURITY.md`, `VERSION`, and this changelog.
  - Standards docs under `docs/standards/` (naming, versioning, commits,
    changelog, Go 1.26 style, Terraform-provider best practices,
    Terragrunt integration) and a chosen delivery methodology
    (`docs/methodology.md`).

### Changed

- **Dependencies bumped to latest (July 2026):** Go directive `1.26.5`;
  `terraform-plugin-framework` `v1.13.0` → `v1.19.0`; `grpc` `v1.78.0` →
  `v1.82.1` (clears the GO-2026-4762 advisory noted in the v0.1 skeleton);
  `protobuf` `v1.36.11`. Added `terraform-plugin-testing v1.16.0`,
  `-validators v0.19.0`, `-log v0.10.0`, `-go v0.31.0` to the module graph.

### Removed

- **`aristanetworks/cloudvision-go` dependency dropped for good.** It dragged a
  conflicting ancient `google.golang.org/genproto` (broke `go mod tidy`) and
  pulled the entire CloudVision Go surface for the sake of three services.
  Replaced by a **custom buf-generated gRPC client** (ADR 0005): vendored protos
  under `api/proto/`, generated stubs in `internal/pb/` via `buf`
  (`buf.yaml` + `buf.gen.yaml`, pinned `protoc-gen-go` v1.36.11 /
  `protoc-gen-go-grpc` v1.6.2), `make proto` / `proto-lint` / `proto-breaking`,
  and `buf` added to the dev image. Actual stub generation is the first P1
  gRPC-wiring step; the provider builds dependency-free until then.

## [0.1.0] — 2026-07-16

### Added

- Initial `terraform-provider-cvp` v0.1 skeleton: provider server bootstrap,
  provider schema (`endpoint`, `auth_method`, `cert`/`session`/`bearer`), and
  three schema-only resources with stubbed CRUD (`cvp_workspace`,
  `cvp_studio_inputs`, `cvp_change_control`). See `design.md`.

[Unreleased]: https://github.com/ioplane/terraform-provider-cvp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ioplane/terraform-provider-cvp/releases/tag/v0.1.0
