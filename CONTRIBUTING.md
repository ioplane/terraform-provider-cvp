# Contributing

This provider follows the same house standard as the sibling `pulumi-eos` and
`arista-cvp-re` repositories. Read [`AGENTS.md`](AGENTS.md) first — it is the
canonical guide for both human and agent contributors.

## Standards

| Item | Standard |
|---|---|
| Versioning | [SemVer 2.0.0](https://semver.org/spec/v2.0.0.html) — see [`docs/standards/versioning.md`](docs/standards/versioning.md) |
| Commits | [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) — see [`docs/standards/commits.md`](docs/standards/commits.md) |
| Changelog | [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/) — see [`docs/standards/changelog.md`](docs/standards/changelog.md) |
| Naming | [`docs/standards/naming-conventions.md`](docs/standards/naming-conventions.md) |
| Go style | [`docs/standards/go-1.26-style.md`](docs/standards/go-1.26-style.md) |
| Provider design | [`docs/standards/terraform-provider-best-practices.md`](docs/standards/terraform-provider-best-practices.md) |
| Methodology | [`docs/methodology.md`](docs/methodology.md) |

## Development environment

Everything runs inside the Podman dev container — no host Go toolchain is
required. See [`docs/development.md`](docs/development.md).

```bash
task up        # build + start the dev container
task shell     # open a shell inside it
task all       # build + test + lint + tffmt-check + lint-docs + vulncheck
```

## Workflow — worktree + Pull Request only

**`main` is never committed to directly.** Every change is developed on a branch
in an isolated **git worktree** and integrated via a **GitHub Pull Request**
(this project uses PRs, not GitLab MRs). No direct pushes to `main`, no local
merges — branch protection enforces it (see [`GOVERNANCE.md`](GOVERNANCE.md)).

1. **Create a worktree:**

   ```bash
   scripts/worktree.sh new feat/<scope>/<name>     # maintenance
   scripts/worktree.sh new sprint/<id>-<scope>     # sprint work
   ```

   Branch kinds: `feat/<scope>/<name>`, `fix/<scope>/<name>`,
   `chore/<scope>/<name>`, or `sprint/<id>-<scope>` (one branch per sprint).
2. `cd` into the worktree; develop inside the dev container.
3. `task all` must pass before pushing. Resource-touching changes must also pass
   `task verify` (live acceptance tests against the netlab2 CVP lab — see
   [`AGENTS.md`](AGENTS.md)).
4. Commit messages follow Conventional Commits 1.0.0; the scope is one of the
   values in `.commitlintrc.yaml`.
5. Update the `[Unreleased]` block of `CHANGELOG.md` for every user-visible
   change.
6. Push the branch and **open a PR**. The PR **title** is itself a Conventional
   Commit subject (validated in CI). One code-owner approval + green required
   checks → **squash-merge**.
7. `scripts/worktree.sh rm <branch>` to clean up.

## Quality gates

| Gate | Command |
|---|---|
| Build | `task build` |
| Unit tests + race | `task test` |
| Acceptance (live CVP) | `task testacc` |
| Go static analysis (golangci-lint v2) | `task lint` |
| Vulnerabilities | `task vulncheck` · `task osv-scan` |
| Terraform formatting | `task tffmt-check` |
| Registry docs | `task docs-check` |
| Docs lint (md + yaml + spell) | `task lint-docs` |

The lint gate is **zero-findings**: `golangci-lint run` exits non-zero on any
finding, so a PR may not merge while any golangci-lint finding is unresolved.
The `error`/`warning` severity tiers classify findings for triage; they do not
change the exit code.

## Definition of done (per resource)

See [`docs/standards/terraform-provider-best-practices.md`](docs/standards/terraform-provider-best-practices.md)
§ "Definition of done".

## Evidence discipline

Any factual claim about Arista CVP/EOS behaviour in `docs/` or a resource's
package comment must be corroborated through the local `arista-mcp` MCP server
and cited. See `AGENTS.md § Evidence discipline`.
