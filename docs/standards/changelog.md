# Changelog conventions

[`CHANGELOG.md`](../../CHANGELOG.md) follows
[Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/).

## Rules

- **Human-written, curated.** Not a `git log` dump — describe user impact.
- **`[Unreleased]` accumulates** at the top. Every PR with a user-visible change
  adds an entry there (checked in the PR template).
- **Group by change type** in this order: `Added`, `Changed`, `Deprecated`,
  `Removed`, `Fixed`, `Security`.
- **Newest release first.** Each release is `## [X.Y.Z] — YYYY-MM-DD`
  (ISO 8601 date, em-dash).
- **Breaking changes** go under `Changed` with a `BREAKING:` prefix and a
  migration note.
- **Link the diff** at the bottom (`[Unreleased]`, `[X.Y.Z]` compare/tag links).

## Release cut

1. Rename `## [Unreleased]` content into a new dated `## [X.Y.Z] — YYYY-MM-DD`
   section; recreate an empty `[Unreleased]`.
2. Update `VERSION`.
3. Commit `chore(release): X.Y.Z`.
4. Tag `vX.Y.Z` (annotated, GPG-signed).
5. The release workflow `awk`-extracts the section between two `## [` headers
   and feeds it to goreleaser as the GitHub Release notes.

Because step 5 parses the file mechanically, keep the heading format exact:
`## [X.Y.Z] — YYYY-MM-DD` with the version in square brackets.
