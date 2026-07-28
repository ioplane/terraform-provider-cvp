<!--
PR title MUST be a Conventional Commit subject, e.g.:
  feat(workspace): wire Create/Read against WorkspaceConfigService
It is validated by the commitlint CI job.
-->

## Summary

<!-- What and why. Link the relevant issue / sprint exit-criteria. -->

## Changes

-

## Checklist

- [ ] `task all` passes locally (build + test + lint + tffmt-check + lint-docs + vulncheck).
- [ ] Resource-touching changes: `task verify` passes (live CVP acceptance tests) and the `N/N acceptance tests pass` line is quoted in the PR body.
- [ ] `CHANGELOG.md` `[Unreleased]` updated for every user-visible change.
- [ ] Registry docs regenerated (`task docs`) if schema changed.
- [ ] New Arista-behaviour claims are corroborated via `arista-mcp` and cited.
- [ ] No secrets, credentials, or environment identifiers added.
