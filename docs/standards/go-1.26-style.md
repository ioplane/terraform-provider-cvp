# Go 1.26 patterns, antipatterns, standards

Project Go standard for `terraform-provider-cvp`. Pinned to **Go 1.26.5** (dev
container `golang:1.26-trixie`). Aligned with the official
[Go 1.26 release notes](https://go.dev/doc/go1.26) and consistent with the
sibling `pulumi-eos` Go style guide.

Before writing Go, consult the current API via the `gopls` LSP and the
`context7` MCP — do not rely on training-data recall for library signatures.

## 1. Language / stdlib features adopted

| Feature (Go 1.26) | Pattern here | Rationale |
|---|---|---|
| `new(value)` with an expression | `new("Loopback0")` for `*string` schema defaults | Drops the `v := …; &v` ceremony; flagged by `modernize`. |
| `errors.AsType[T]()` | `if e, ok := errors.AsType[*status.Status](err); ok {…}` | Type-safe error matching over gRPC `status`. |
| `slog.NewMultiHandler` | compose sinks if provider logging grows | Provider logging is primarily `tflog`; slog only for CLI-side tooling. |
| `strings.SplitSeq` | line iteration without a `[]string` alloc | `modernize/stringsseq`. |
| `reflect.Type.Fields()/.Methods()` | schema/codegen tooling only | Never in CRUD hot paths. |
| Green Tea GC (default) | no code change | 10–40 % less GC overhead. |
| Heap address randomization (default) | no code change | Mitigates address-prediction attacks. |

## 2. Provider-specific conventions

| Area | Rule |
|---|---|
| Logging | `github.com/hashicorp/terraform-plugin-log/tflog` — **not** `log` or `fmt.Print*`. `depguard` denies `^log$`; `forbidigo` denies `fmt.Print*` outside `main.go`/`scripts/`. |
| Errors surfaced to Terraform | `resp.Diagnostics.AddError(summary, detail)` / `AddAttributeError`. Never `panic`, never a bare returned string. |
| Internal errors | Sentinel `var ErrXxx = errors.New("…")` at package level; wrap with `fmt.Errorf("…: %w", err)` (`err113`, `errorlint`). |
| Context | `context.Context` is the first argument of every RPC/CRUD helper; honour `ctx.Done()`; never store a ctx in a struct (`containedctx`). |
| Pointer receivers | Every method on a resource type uses a pointer receiver (`recvcheck`). |
| Framework types | Convert `types.String`/`types.Bool` at the boundary; never assume non-null — check `IsNull()`/`IsUnknown()`. |
| Randomness | `crypto/rand` for any secret/token; `math/rand/v2` otherwise. `depguard` denies `math/rand`. |
| TLS | Minimum TLS 1.3 for the CVP client; leave the Go 1.26 post-quantum hybrid defaults on; `InsecureSkipVerify` only behind the documented `insecure_tls` provider flag, never a code default. |

## 3. Antipatterns banned

| Antipattern | Replacement | Enforced by |
|---|---|---|
| `v := "x"; p := &v` | `p := new("x")` | `modernize/newexpr` |
| `var e *E; errors.As(err, &e)` | `errors.AsType[*E](err)` | manual review |
| `log.Printf` / `fmt.Println` in library code | `tflog.Debug/Info/Error` | `depguard`, `forbidigo` |
| `panic(...)` in CRUD | diagnostic + early return | `gocritic` |
| Naked returns in non-trivial funcs | explicit returns | `nakedret max-func-lines: 0` |
| `math/rand` for tokens | `crypto/rand` | `depguard`, `gosec` |
| `tls.Config{InsecureSkipVerify: true}` as default | CA bundle via `RootCAs` | `gosec` |
| `interface{}` where a concrete type fits | concrete type / small interface | `iface`, `interfacebloat` |

## 4. Standards (mandatory)

| Area | Standard |
|---|---|
| Module path | `github.com/ioplane/terraform-provider-cvp`. |
| Go directive | `go 1.26.5`. |
| Layout | `internal/provider`, `internal/resources/<area>`, `internal/client/cvp`. Framework resources are unexported in `internal/`. |
| Imports | `gofumpt` order; `goimports.local-prefixes` set to the module path. |
| Test names | `Test<Subject>_<Behaviour>`; acceptance tests `TestAcc<Resource>_<Case>`. |
| Test parallelism | `t.Parallel()` mandatory in **unit** tests; **forbidden** in acceptance tests (shared live CVP). Enforced by `paralleltest` with a `test/acceptance/` exclusion. |
| Race detector | `-race` always on in `make test` and CI. |
| Build tags | `//go:build acceptance` for live-CVP tests. |
| Vulnerability gate | `govulncheck` + `osv-scanner`; any allow-listed advisory needs a documented reason. |

## 5. Tooling adopted

| Tool | Version | Use |
|---|---|---|
| `go` | 1.26.5 | language / build |
| `gofmt` / `gofumpt` / `goimports` | bundled / latest | format gate |
| `golangci-lint` | v2.12.2 | aggregate linter, allowlist mode (`.golangci.yml`) |
| `gopls` | latest | LSP: references / rename / find-symbol / diagnostics |
| `govulncheck` | v1.6.0 | vulnerability gate |
| `osv-scanner` | v2.4.0 | vulnerability gate |
| `tfplugindocs` | v0.25.0 | registry doc generation |
| `gotestsum` | latest | JUnit test reporting in CI |

## 6. Reading list

- [Go 1.26 release notes](https://go.dev/doc/go1.26) — authoritative.
- [Effective Go](https://go.dev/doc/effective_go) — baseline style.
- [Terraform Plugin Framework docs](https://developer.hashicorp.com/terraform/plugin/framework)
  — the framework's own conventions win where this guide is silent.
