<div align="center">

# Naming conventions

</div>

Consistent names make the repository searchable, sortable, and legible to both
humans and agents. This standard adapts two general references —
[Harvard HMS file-naming conventions](https://datamanagement.hms.harvard.edu/plan-design/file-naming-conventions)
and the [IT Glue naming best-practices](https://www.itglue.com/blog/naming-conventions-examples-formats-best-practices/) —
to a Terraform provider codebase.

## 1. Universal rules

Derived from both references:

- **Alphanumeric + `-` + `_` only.** No spaces, no special characters
  (`~ ! @ # $ % ^ & * ( ) : < > ? , [ ] { } ' " |`).
- **One separator per context** (see the case table below); never mix within a
  single identifier.
- **ISO 8601 dates** (`YYYY-MM-DD`), so names sort chronologically. Timestamps:
  `YYYY-MM-DDThhmm`.
- **Most significant token first**, so default lexical sort is useful.
- **Concise but descriptive.** Aim ≤ 40–50 characters for file names.
- **No vague descriptors** (`final`, `new`, `latest`, `copy`) — that is what
  Git and SemVer are for.
- **No undocumented abbreviations.** If you coin one, add it to the glossary in
  this file and to `.cspell.json`.

## 2. Case by context

| Context | Case | Example |
|---|---|---|
| Go files | `snake_case.go` | `change_control.go`, `provider_test.go` |
| Go packages | short, lowercase, no underscores | `changecontrol`, `studioinputs` |
| Go exported identifiers | `PascalCase` | `WorkspaceResource`, `ClientBundle` |
| Go unexported identifiers | `camelCase` | `buildTLSCreds`, `providerModel` |
| Markdown / docs | `kebab-case.md` | `naming-conventions.md`, `go-1.26-style.md` |
| Dated notes / ADRs | `NNNN-kebab.md` / `YYYY-MM-DD-kebab.md` | `0001-methodology.md` |
| Directories | `kebab-case` or single word | `deployments/`, `internal/client/` |
| YAML / config | `kebab-case.yml` | `compose.dev.yml`, `dependabot.yml` |
| Terraform files | `snake_case.tf` | `main.tf`, `variables.tf` |
| Env vars | `SCREAMING_SNAKE_CASE` | `CVP_ENDPOINT`, `TF_ACC` |

## 3. Terraform-facing names (the contract)

These names are the **public API** of the provider. Changing one is a breaking
change (see [`versioning.md`](versioning.md)). They follow HashiCorp's own
convention, which mirrors the underlying CloudVision API.

- **Resource / data-source type:** `cvp_<noun>`, `snake_case`, singular.
  - `cvp_workspace`, `cvp_studio_inputs`, `cvp_change_control`,
    `cvp_configlet`, `cvp_configlet_assignment`, `cvp_tag_assignment`.
- **Attributes:** `snake_case`, matching the CVP field where practical.
  - `display_name`, `auto_approve`, `wait_for_execution`, `workspace_id`.
- **Boolean attributes describe an action, default `false`** (HashiCorp design
  principle): `auto_build`, `stop_on_failure`, `rollback_on_failure` — not
  `disable_*`.
- **IDs are `id`** (computed), with typed references elsewhere
  (`workspace_id`, `studio_id`).
- **Timestamps:** RFC 3339 strings; suffix `_at` (`created_at`, `started_at`).

## 4. Git branches

| Kind | Pattern | Example |
|---|---|---|
| Sprint | `sprint/<sprint-id>-<scope>` | `sprint/S2-workspace-crud` |
| Feature | `feat/<scope>/<name>` | `feat/studio/inputs-batch-set` |
| Fix | `fix/<scope>/<name>` | `fix/auth/bearer-token-refresh` |
| Chore | `chore/<scope>/<name>` | `chore/deps/bump-plugin-framework` |

`<scope>` is one of the commit scopes enumerated in `.commitlintrc.yaml`.

## 5. Commits, versions, tags

- Commit subjects: [Conventional Commits](commits.md).
- Versions: [SemVer 2.0.0](versioning.md); the single source of truth is
  `VERSION`.
- Tags: `vX.Y.Z` (annotated, GPG-signed).
- Release archives (goreleaser): `terraform-provider-cvp_<version>_<os>_<arch>.zip`
  — mandated verbatim by the Terraform Registry.

## 6. Glossary of accepted abbreviations

| Abbrev | Meaning |
|---|---|
| `cvp` | CloudVision Portal |
| `cvaas` | CloudVision as a Service |
| `cc` | change control |
| `mtls` | mutual TLS |
| `acc` | acceptance (test) |
| `crud` | create/read/update/delete |

Add new entries here and to `.cspell.json` in the same commit that introduces
them.
