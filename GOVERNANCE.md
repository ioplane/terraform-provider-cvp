# Governance

How decisions are made in `terraform-provider-cvp`. This complements the
delivery method in [`docs/methodology.md`](docs/methodology.md).

## Roles

| Role | Responsibility |
|---|---|
| **Maintainers** (`@ioplane/maintainers`) | Own the public contract (schema/state/registry), review and merge PRs, cut releases, triage security reports. Listed in [`.github/CODEOWNERS`](.github/CODEOWNERS). |
| **Contributors** | Anyone opening a PR. Bound by [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) and [`CONTRIBUTING.md`](CONTRIBUTING.md). |

## Decision making

- **Code / docs changes:** via Pull Request with at least one maintainer
  (code-owner) approval and green required checks. No direct pushes to `main`.
- **Architecture / contract changes:** recorded as an
  [ADR](docs/adr/) and frozen at the design gate for a minor
  (see methodology). Breaking changes follow [SemVer](docs/standards/versioning.md).
- **Releases:** a maintainer cuts the version per [`docs/release.md`](docs/release.md).
- **Disagreement:** maintainers seek consensus; if none, the change is deferred
  and the discussion recorded in the PR/ADR.

## Branch protection (required on `main`)

- Require a PR before merging; no direct pushes.
- Require code-owner review.
- Require status checks: CI (build, lint-go, vulncheck, terraform, lint-docs),
  Security, and commitlint (PR title) to pass.
- Require branches to be up to date before merging.
- Require signed commits (recommended).
- Restrict force-pushes and deletions.

## Changes to this document

Via PR, maintainer approval required.
