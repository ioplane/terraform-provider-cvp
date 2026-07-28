package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	workspaceactions "github.com/ioplane/terraform-provider-cvp/internal/actions/workspace"
	"github.com/ioplane/terraform-provider-cvp/internal/client/cvp"
	changecontrolres "github.com/ioplane/terraform-provider-cvp/internal/resources/change_control"
	inputsres "github.com/ioplane/terraform-provider-cvp/internal/resources/studio_inputs"
	workspaceres "github.com/ioplane/terraform-provider-cvp/internal/resources/workspace"
)

var (
	_ provider.Provider            = &cvpProvider{}
	_ provider.ProviderWithActions = &cvpProvider{}
)

type cvpProvider struct {
	version string
}

// New returns a provider factory bound to the given build version. The version
// is surfaced through Metadata and injected at build time via -ldflags.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &cvpProvider{version: version}
	}
}

type providerModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	AuthMethod  types.String `tfsdk:"auth_method"`
	Token       types.String `tfsdk:"token"`
	CertPEM     types.String `tfsdk:"cert_pem"`
	KeyPEM      types.String `tfsdk:"key_pem"`
	CAPEM       types.String `tfsdk:"ca_pem"`
	InsecureTLS types.Bool   `tfsdk:"insecure_tls"`
}

func (p *cvpProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cvp"
	resp.Version = p.version
}

func (p *cvpProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Arista CloudVision Portal provider. Prototype.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "CVP gRPC endpoint (e.g. `um-cvp01.infra4.dev:443`).",
				Required:            true,
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "One of `session`, `cert`, `bearer`.",
				Required:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Session cookie or bearer token. Required if auth_method=session|bearer.",
				Optional:            true,
				Sensitive:           true,
			},
			"cert_pem": schema.StringAttribute{
				MarkdownDescription: "Client cert PEM. Required if auth_method=cert.",
				Optional:            true,
			},
			"key_pem": schema.StringAttribute{
				MarkdownDescription: "Client key PEM. Required if auth_method=cert.",
				Optional:            true,
				Sensitive:           true,
			},
			"ca_pem": schema.StringAttribute{
				MarkdownDescription: "Trust root PEM. Optional (defaults to system trust store).",
				Optional:            true,
			},
			"insecure_tls": schema.BoolAttribute{
				MarkdownDescription: "Skip server cert verification. **Not for production.**",
				Optional:            true,
			},
		},
	}
}

func (p *cvpProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If any value is unknown (a reference to a not-yet-applied resource),
	// ValueString/ValueBool would collapse it to a zero value and build a
	// mis-configured client. Refuse to configure until the value is known.
	for _, u := range []struct {
		attr    string
		unknown bool
	}{
		{"endpoint", cfg.Endpoint.IsUnknown()},
		{"auth_method", cfg.AuthMethod.IsUnknown()},
		{"token", cfg.Token.IsUnknown()},
		{"cert_pem", cfg.CertPEM.IsUnknown()},
		{"key_pem", cfg.KeyPEM.IsUnknown()},
		{"ca_pem", cfg.CAPEM.IsUnknown()},
		{"insecure_tls", cfg.InsecureTLS.IsUnknown()},
	} {
		if u.unknown {
			resp.Diagnostics.AddAttributeError(path.Root(u.attr),
				"Unknown provider configuration value",
				"The CVP provider cannot be configured while "+u.attr+" is unknown. "+
					"Set it to a static value or apply the resource it references first.")
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := cvp.New(cvp.Config{
		Endpoint:    cfg.Endpoint.ValueString(),
		AuthMethod:  cfg.AuthMethod.ValueString(),
		Token:       cfg.Token.ValueString(),
		CertPEM:     []byte(cfg.CertPEM.ValueString()),
		KeyPEM:      []byte(cfg.KeyPEM.ValueString()),
		CAPEM:       []byte(cfg.CAPEM.ValueString()),
		InsecureTLS: cfg.InsecureTLS.ValueBool(),
	})
	if err != nil {
		mapClientError(err, resp)
		return
	}

	// The lazy gRPC client is shared with every resource, data source and
	// action via ProviderData; each type-asserts it to *cvp.Client in Configure.
	resp.DataSourceData = client
	resp.ResourceData = client
	resp.ActionData = client
}

// mapClientError turns cvp client-construction errors into actionable,
// attribute-scoped diagnostics.
func mapClientError(err error, resp *provider.ConfigureResponse) {
	switch {
	case errors.Is(err, cvp.ErrEndpointRequired):
		resp.Diagnostics.AddAttributeError(path.Root("endpoint"), "endpoint is required", err.Error())
	case errors.Is(err, cvp.ErrUnknownAuthMethod):
		resp.Diagnostics.AddAttributeError(path.Root("auth_method"),
			"invalid auth_method", "auth_method must be one of cert, session or bearer.")
	case errors.Is(err, cvp.ErrTokenRequired):
		resp.Diagnostics.AddAttributeError(path.Root("token"),
			"token is required", "auth_method session/bearer requires token.")
	case errors.Is(err, cvp.ErrClientCertRequired):
		resp.Diagnostics.AddAttributeError(path.Root("cert_pem"),
			"client certificate is required", "auth_method cert requires cert_pem and key_pem.")
	default:
		resp.Diagnostics.AddError("CVP client initialisation failed", err.Error())
	}
}

func (p *cvpProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		workspaceres.NewResource,
		inputsres.NewResource,
		changecontrolres.NewResource,
	}
}

func (p *cvpProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

// Actions exposes the CVP workspace workflow verbs as Terraform actions
// (ProviderWithActions, Terraform >= 1.14). See ADR 0006.
func (p *cvpProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{
		workspaceactions.NewBuildAction,
		workspaceactions.NewCancelBuildAction,
		workspaceactions.NewSubmitAction,
		workspaceactions.NewAbandonAction,
		workspaceactions.NewRollbackAction,
		workspaceactions.NewRebaseAction,
	}
}
