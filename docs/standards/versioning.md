<div align="center">

# Versioning

</div>

The provider follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).
The provider binary, its JSON schema, and (when published) its docs share one
version. The single source of truth is the [`VERSION`](../../VERSION) file; the
Git tag `vX.Y.Z` is cut from it at release time.

## What "the public API" means for a provider

A change is **breaking** if an unmodified, previously-valid Terraform
configuration or state would newly error, silently change meaning, or force
replacement. Concretely, breaking changes include:

- removing or renaming a resource, data source, or attribute;
- making an optional attribute required, or removing a default;
- narrowing an attribute's accepted values or type;
- changing an attribute from updatable to `RequiresReplace`;
- changing the shape of `id` or of imported state;
- raising the minimum Terraform or provider-protocol version.

Non-breaking (minor/patch): adding a resource, adding an optional attribute,
adding computed output, widening accepted input, bug fixes that restore
documented behaviour.

## Rules

| Component | Bumps when |
|---|---|
| MAJOR (`X`) | Breaking change — **after `1.0.0`**. |
| MINOR (`Y`) | Backward-compatible feature; **also breaking changes while `0.x`** (SemVer §4). |
| PATCH (`Z`) | Backward-compatible bug fix. |

## Pre-1.0 policy

While `0.x`, the schema may still change between minors. Per SemVer §4, anything
MAY change in `0.x`; this project constrains that to: breaking changes bump the
**MINOR** and are flagged in `CHANGELOG.md` under `Changed` with a `BREAKING:`
prefix and a migration note. Patch releases never break.

## Pre-releases

`vX.Y.Z-rc.N` for release candidates, `-alpha.N` / `-beta.N` earlier. A
pre-release tag never becomes a stable GitHub Release or a Terraform Registry
version; it is validated against the netlab2 CVP lab first.

## State-upgrade obligation

Any schema change that alters stored state shape must ship a
`ResourceWithUpgradeState` implementation so existing state migrates cleanly.
This is part of the per-resource Definition of Done.

## Dependency version policy

"Always latest, pinned." The intent is to track the newest releases (Go,
Terraform Plugin Framework, gRPC, tooling, GitHub Actions) while pinning exact
versions for reproducibility:

- `go.mod` `go` directive: exact patch (`1.26.5`).
- Go dependencies: exact tags; Dependabot proposes weekly bumps.
- Container tool versions: pinned as `ARG`s in `Containerfile.dev` and mirrored
  in `compose.dev.yml`.
- GitHub Actions: pinned by commit SHA with a `# vX.Y.Z` trailing comment.
- Base image `golang:1.26-trixie` intentionally floats within the `1.26`
  minor so patch releases (currently `1.26.5`) are picked up on rebuild.
