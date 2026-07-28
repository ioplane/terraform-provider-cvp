# Support

Thanks for using `terraform-provider-cvp`. Here is where to go:

| I want to… | Go to |
|---|---|
| Ask a usage question | [GitHub Discussions](https://github.com/ioplane/terraform-provider-cvp/discussions) |
| Report a bug | [New issue → Bug report](https://github.com/ioplane/terraform-provider-cvp/issues/new/choose) |
| Request a feature | [New issue → Feature request](https://github.com/ioplane/terraform-provider-cvp/issues/new/choose) |
| Report a security vulnerability | **Privately** via [Security advisories](https://github.com/ioplane/terraform-provider-cvp/security/advisories/new) — see [`SECURITY.md`](SECURITY.md) |
| Contribute | [`CONTRIBUTING.md`](CONTRIBUTING.md) + [`AGENTS.md`](AGENTS.md) |

## Before opening an issue

- Reproduce on the latest released version.
- Include provider, Terraform/OpenTofu, and CVP versions.
- Provide a minimal HCL reproduction.
- **Never** paste secrets, tokens, certificates, Terraform state, or CVP support
  bundles. Redact endpoints and identifiers.

## Scope

This provider targets **on-prem CloudVision Portal**. CVaaS-only shapes are a
documented non-goal (see [`design.md`](design.md)). Questions about Arista EOS
itself belong upstream with Arista.
