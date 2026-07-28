<div align="center">

# Vendored CloudVision API protobufs

</div>

This directory holds the **vendored** Protocol Buffer definitions for the subset
of the Arista CloudVision Resource APIs that `terraform-provider-cvp` consumes.
From them we generate our own Go stubs into [`internal/pb`](../../internal/pb)
via `buf` — we do **not** depend on `github.com/aristanetworks/cloudvision-go`.
See [ADR 0005](../../docs/adr/0005-custom-cvp-client.md) for the rationale.

## Source

Upstream: <https://github.com/aristanetworks/cloudvision-apis> (Apache-2.0).

Keep only what the provider calls, plus transitive imports:

```text
api/proto/
├── arista/
│   ├── workspace.v1/        # WorkspaceConfigService, WorkspaceService
│   ├── studio.v1/           # InputsConfigService, InputsService, SecretInput*
│   └── changecontrol.v1/    # ChangeControlConfigService, ApproveConfigService, …
├── fmp/                     # Arista field-mask / wrapper helpers (imported)
└── google/                  # only if not resolved from buf's well-known types
```

## Updating the vendored protos

1. Pull the pinned upstream revision (record the commit in the PR body).
2. Copy only the needed service files + their imports under `api/proto/`.
3. `task proto-lint` — buf lint.
4. `task proto-breaking` — buf breaking-change check against the previous revision.
5. `task proto` — regenerate `internal/pb`.
6. `task build test` — confirm the generated client still compiles and passes.

Pin the upstream revision in the commit message; treat a proto bump like any
other dependency bump (`build(deps)` commit, CHANGELOG entry).

## Current vendored closure

Vendored from `cloudvision-apis` — the `arista/workspace.v1` and
`arista/studio.v1` import closures (computed with
`buf build … | buf ls-files --include-imports`); Google well-known types are
provided by buf and not vendored:

```text
arista/workspace.v1/{workspace,services.gen}.proto
arista/studio.v1/{studio,services.gen}.proto
arista/{configstatus.v1,imagestatus.v1,subscriptions,time}/*.proto
fmp/{deletes,extensions,wrappers}.proto
```

`studio.v1` adds no new transitive imports beyond the workspace closure
(`fmp/*`, `subscriptions`, `time`, Google well-known types).

The pinned upstream revision is recorded in `.upstream-revision`. `buf lint`
skips these vendored trees (see `buf.yaml`); `task proto` regenerates
`internal/pb` deterministically.
