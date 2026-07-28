# ADR 0001 — Delivery methodology

- **Status:** Accepted
- **Date:** 2026-07-28

## Context

The provider has a hard external contract (schema/state/registry/SemVer) yet a
large, incrementally-deliverable resource surface against a partially
reverse-engineered target (CVP gRPC). We need a method that protects the
contract without stalling per-resource discovery, and that matches the sibling
repos (`pulumi-eos`, `arista-cvp-re`).

## Decision

Adopt **gated-iterative delivery**: Waterfall macro-phases (requirements →
design → implementation → verification → release → maintenance) with signed exit
gates, iterative 2-week sprints inside implementation, Trunk-Based Development, a
per-resource Definition of Done, and `arista-cvp-re`-style evidence discipline.
Full description in [`../methodology.md`](../methodology.md).

## Consequences

- Contract-affecting work is frozen at the design gate per minor, reducing
  breaking churn.
- Progress is measured in resources meeting the DoD, not lines of code.
- Every CVP-behaviour claim is corroborated via `arista-mcp` + the lab, keeping
  a reverse-engineered target honest.
- Slightly heavier up-front design than pure Agile — accepted, because schema
  breakage is expensive post-release.
