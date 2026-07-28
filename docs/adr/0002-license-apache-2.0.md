# ADR 0002 — License: Apache-2.0

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

The provider needs a license compatible with the Terraform Registry, with the
Arista sources it derives from (the vendored `cloudvision-apis` protos are
Apache-2.0; see ADR 0005), and with the sibling repositories.

## Decision

License under **Apache-2.0**.

## Consequences

- Matches the vendored `cloudvision-apis` protos (Apache-2.0) and the sibling
  `pulumi-eos` (ADR-2 there also chose Apache-2.0) — one license story across
  the house.
- Accepted by the Terraform Registry.
- Provides an explicit patent grant (unlike MIT), appropriate for a
  network-infrastructure provider.
- MPL-2.0 (HashiCorp's historical provider license) was considered but rejected
  for Arista-library alignment and house consistency.
