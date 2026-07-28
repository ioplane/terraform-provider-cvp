<div align="center">

# Security Policy

</div>

## Supported versions

While the provider is pre-`1.0.0`, only the latest `0.x` minor receives
security fixes. After `1.0.0`, the two most recent minor versions are
supported.

## Reporting a vulnerability

**Do not open a public issue for a security vulnerability.**

Report privately via GitHub's [private vulnerability reporting](https://github.com/ioplane/terraform-provider-cvp/security/advisories/new)
("Report a vulnerability" on the Security tab). Include:

- affected version(s) and platform,
- a description and, where possible, a minimal reproduction,
- the impact you have assessed.

You will get an acknowledgement within **3 business days** and a remediation
plan within **10 business days**. Critical, actively-exploitable issues are
patched within a **7-day SLA**.

## Handling secrets

This provider talks to Arista CloudVision over authenticated gRPC/TLS. Observe
the following:

- **Never commit credentials.** CVP tokens, client certificates/keys, session
  cookies, and lab passwords live in `gopass` (see `AGENTS.md`), never in the
  repository, examples, or test fixtures.
- **`token`, `key_pem` provider attributes are `Sensitive`** and are redacted
  in Terraform output — but they are still persisted to state. Treat the
  Terraform state as a secret: use an encrypted backend and restrict access.
- **Prefer `bearer` (rotatable API token) for CI/CD.** Session cookies and
  long-lived certificates are harder to rotate.
- **`insecure_tls` is not for production.** It disables server-certificate
  verification and exists only for lab bring-up.
- Studio `is_secret` inputs are never read back by the provider (drift
  detection is disabled for them by design); reference an external secret
  manager for their values.

## Supply chain

- Releases ship a GPG-signed `SHA256SUMS` and an SPDX SBOM.
- Dependencies are scanned by `govulncheck`, `osv-scanner`, and Trivy; CodeQL
  and `gosec` run on every push and weekly; OpenSSF Scorecard runs weekly.
- All GitHub Actions are pinned by commit SHA.
