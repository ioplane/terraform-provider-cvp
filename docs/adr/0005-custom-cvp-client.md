<div align="center">

# ADR 0005 — Custom buf-generated CVP client (drop cloudvision-go)

</div>

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

The provider talks to CloudVision Portal over gRPC. The obvious dependency,
`github.com/aristanetworks/cloudvision-go`, has two problems for us:

1. It transitively requires an **ancient monolithic
   `google.golang.org/genproto`** that collides with the modern split
   `genproto/googleapis/rpc` module and breaks `go mod tidy` (this is why the
   require was removed from `go.mod` at repo bootstrap).
2. It pulls the **entire** CloudVision Go surface when the provider only needs a
   handful of services (`workspace.v1`, `studio.v1`, `changecontrol.v1`), and we
   do not control its dependency versions.

## Decision

Generate our **own** Go gRPC stubs from the vendored CloudVision `.proto`
definitions using **buf**, instead of depending on `cloudvision-go`.

- Vendored protos: `api/proto/` (subset of `aristanetworks/cloudvision-apis`,
  Apache-2.0, pinned per update).
- Generation: `buf` (`buf.yaml` + `buf.gen.yaml`) with pinned
  `protoc-gen-go` (v1.36.11) and `protoc-gen-go-grpc` (v1.6.2), matching
  `go.mod`'s `protobuf`/`grpc`.
- Output: `internal/pb/` — generated, `DO NOT EDIT`, excluded from all linters,
  spell-check, and code review.
- Hand-written client wrapper (retry, auth, helpers) stays in
  `internal/client/cvp/` and imports `internal/pb`.
- Task targets: `task proto` (generate), `task proto-lint`, `task proto-breaking`.

## Consequences

- **Full version control:** the generated code uses our exact `protobuf`/`grpc`,
  with a clean `genproto` — no ambiguous-import breakage.
- **Minimal surface:** only the services we call are generated, keeping the
  module graph small.
- **Reproducible:** pinned plugins + vendored protos → deterministic output; a
  proto bump is a normal `build(deps)` change gated by `buf breaking`.
- **Generated noise is contained:** `internal/pb` is machine-owned and hidden
  from review/lint/diff tooling.
- **Cost:** we own the vendoring + regeneration step. Accepted — it is a
  mechanical `task proto` and buys us independence from `cloudvision-go`'s
  dependency choices.
- If Arista later ships a `cloudvision-go` release with a clean `genproto`, this
  decision can be revisited; the `internal/client/cvp` wrapper isolates the
  provider from the stub source either way.

## Alternatives rejected

- **Fork `cloudvision-go`** and `replace` it: still carries the whole library
  and an ongoing upstream-merge burden.
- **Hand-written gRPC client** with no generation: loses compile-time type
  safety, requires hand marshalling of large messages, and silently drifts from
  the server contract.
