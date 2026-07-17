# terraform-provider-cvp

Prototype Terraform provider для Arista CloudVision Portal 2026.2.0.
Scope: **Tier 1 resources** из [I1 declarable-state catalog](../../notes/2026-07-16-i1-declarable-state.md).

**Status**: skeleton только — schema authored, CRUD hooks stubbed with `TODO(P1)`.
См. [design](./design.md) для сравнения с terraform-provider-cvaas + release plan.

## Resources в v0.1 (this prototype)

| Resource | Backing service | State |
|---|---|---|
| `cvp_workspace` | `arista.workspace.v1.WorkspaceConfigService` | Schema только |
| `cvp_studio_inputs` | `arista.studio.v1.InputsConfigService` | Schema + JSON validation |
| `cvp_change_control` | `arista.changecontrol.v1.*` | Schema только |

## Planned для v0.2

- `cvp_configlet` + `cvp_configlet_assignment`
- `cvp_tag_assignment` (v2)
- Full CRUD wiring для v0.1 resources
- Retry + exponential backoff
- `cvp_workspace_build`, `cvp_workspace_submit`, `cvp_workspace_approval` — decoupled workflow steps
- Import state support для всех resources

## Планируется для v0.3

- AAA: `cvp_service_account`, `cvp_service_account_token`, `cvp_api_token`
- Auth: `cvp_saml_provider`, `cvp_oauth_provider`
- License: `cvp_license_file`, `cvp_license_assignment`

## Build

```bash
cd docs/integration-drafts/terraform-provider-cvp/
go build ./...
```

**Note**: prototype ссылается на `cloudvision-go` proto stubs, но `google.golang.org/grpc v1.78.0`
имеет GO-2026-4762 vulnerability (см. arista-cvp-re/docs/notes/2026-07-16-p21-vuln.md).
Production build должен pin к patched >=1.78.1 или actively-maintained fork.

## Auth methods

Три варианта (см. I3 `03-authz-model.md`):

- `cert` — client mTLS с aerisadmin.crt / user cert / package-cert.
- `session` — session cookie от `/cvpservice/login/authenticate.do`.
- `bearer` — API token via `arista.arista_portal.v1.APITokenConfigService`.

Для CI/CD — `bearer` recommended (rotatable).

## Layout

```
terraform-provider-cvp/
├── main.go                         # provider server bootstrap
├── go.mod
├── internal/
│   ├── provider/
│   │   └── provider.go             # provider Metadata/Schema/Configure
│   └── resources/
│       ├── workspace/
│       │   └── resource.go         # cvp_workspace
│       ├── studio_inputs/
│       │   └── resource.go         # cvp_studio_inputs
│       └── change_control/
│           └── resource.go         # cvp_change_control
├── examples/
│   └── basic/main.tf               # Working HCL example
├── README.md
└── design.md
```

## Как перенести в отдельный repo

```bash
mkdir /opt/projects/repositories/terraform-provider-cvp
cp -a docs/integration-drafts/terraform-provider-cvp/. \
      /opt/projects/repositories/terraform-provider-cvp/
cd /opt/projects/repositories/terraform-provider-cvp
git init
gh repo create ioplane/terraform-provider-cvp --private --source=. --push
```

Затем провайдер регистрируется в um-mgt через `required_providers.cvp.source = "ioplane/cvp"`
(private registry pattern через locally-installed provider path OR через
`dev_overrides` в `.terraformrc`).

## Cross-refs

- I1 catalog: `docs/notes/2026-07-16-i1-declarable-state.md`
- I2 Studios: `docs/notes/2026-07-16-i2-studios-reverse.md`
- I3 bootstrap: `docs/notes/2026-07-16-i3-cvp-bootstrap.md`
- Workflow: `docs/integration-drafts/um-docs/07-architecture/cvp/06-workspace-changecontrol.md`
- Terragrunt naming: `docs/integration-drafts/um-docs/06-standards/naming-cvp.md`
