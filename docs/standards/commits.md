<div align="center">

# Commit conventions

</div>

Commits follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/),
enforced by `commitlint` (config: [`.commitlintrc.yaml`](../../.commitlintrc.yaml))
in the `commit-msg` pre-commit hook and on PR titles in CI.

## Shape

```text
<type>(<scope>): <summary, imperative, ≤ 72 chars, sentence/lower case>

<body — WHY this change; wrap at 72; blank line before>

<footer — issue refs, BREAKING CHANGE:, Co-Authored-By:>
```

## Types

| Type | Use |
|---|---|
| `feat` | New user-facing capability (resource, attribute, auth method). |
| `fix` | Bug fix. |
| `docs` | Documentation only. |
| `chore` | Repo hygiene / non-product files. |
| `build` | Build system: `Containerfile`, `Taskfile.yml`, `compose`, `go.mod`. |
| `ci` | GitHub Actions / workflows. |
| `refactor` | Behaviour-preserving code change. |
| `perf` | Performance change. |
| `test` | Tests / acceptance harness only. |
| `style` | Formatting only. |
| `revert` | Revert of a prior commit. |

## Scopes

Enumerated in `.commitlintrc.yaml` — one of: `provider`, `workspace`, `studio`,
`configlet`, `changecontrol`, `tag`, `device`, `aaa`, `client`, `auth`, `docs`,
`examples`, `ci`, `build`, `deps`, `release`, `test`, `lint`, `repo`.

## Breaking changes

Either `feat(scope)!: …` / `fix(scope)!: …`, **or** a footer line:

```text
BREAKING CHANGE: <what breaks and the migration path>
```

Every breaking change also gets a `CHANGELOG.md` entry under `Changed` with a
`BREAKING:` prefix. Version impact per [`versioning.md`](versioning.md).

## Examples

```text
feat(workspace): wire Create/Read against WorkspaceConfigService

Implements the first real gRPC round-trip for cvp_workspace, replacing the
TODO(P1) stub. Adds retry with exponential backoff on Unavailable.

3/3 acceptance tests pass against netlab2 um-cvp01.
```

```text
fix(auth): reject session auth_method without a token before dialing

docs(standards): add Terragrunt integration guide

build(deps): bump terraform-plugin-framework to v1.19.0

chore(release): 0.2.0
```

## Trailers

- Co-authorship: `Co-Authored-By: Name <email>`.
- Reference issues in the footer (`Refs: #12`, `Closes: #12`).
- Commits, code, and documentation carry **no tooling/authorship attribution**
  beyond the human author(s).
