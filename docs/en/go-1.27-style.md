# Go 1.27 development standard

Use Go 1.27.3 from the digest-pinned development image and `go 1.27.3` in
`go.mod`. `GOTOOLCHAIN=local` prevents an unnoticed compiler download in the
container. CI reads `go.mod` or uses the same exact base image.

Keep existing provider conventions: context first, framework diagnostics,
`tflog`, TLS verification, crypto/rand identifiers, concrete types, and race
checks. The [Go 1.26 guide](../standards/go-1.26-style.md) records earlier
conventions; its tool versions are historical.

No new language feature is required for the SDK migration. Go 1.27 `go mod
tidy` combines duplicate require blocks; commit the actual tidy result. The
existing schema, state, import IDs and HCL remain the compatibility contract.
The framework stays at its existing pinned version during this migration.

Use golangci-lint v2.15.0 for Go 1.27 analysis. Run `task all` in the development
container. Live acceptance and release publication remain separate gates.

References: [Go 1.27 release notes](https://go.dev/doc/go1.27),
[Go toolchains](https://go.dev/doc/toolchain),
[golangci-lint v2.15.0](https://github.com/golangci/golangci-lint/releases/tag/v2.15.0).
