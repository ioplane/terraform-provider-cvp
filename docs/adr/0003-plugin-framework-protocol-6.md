<div align="center">

# ADR 0003 — Terraform Plugin Framework, protocol 6

</div>

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

A Terraform provider can be built on the legacy SDKv2 or the modern
`terraform-plugin-framework`. The provider needs plan-time validation,
cross-attribute config validation, custom plan modification (immutable studios),
state upgraders, and modern features (write-only attributes, ephemeral
resources) to model CVP cleanly.

## Decision

Build on **`terraform-plugin-framework` v1.19+** targeting **protocol 6.0**
(`terraform-registry-manifest.json` declares `["6.0"]`).

## Consequences

- Access to `ResourceWithModifyPlan`, `ConfigValidators`, `UpgradeState`,
  write-only attributes (Terraform 1.11+), and ephemeral resources — all needed
  for the CVP model (`design.md` D1/D5/D6).
- Requires Terraform ≥ the framework's protocol-6 baseline; older Terraform is
  unsupported (documented as a minimum-version requirement).
- Not compatible with SDKv2 resource code; there is no SDKv2 legacy to migrate.
- If a future need arises to host both protocols, `terraform-plugin-mux` is the
  escape hatch (already in the resolved module graph).
