package provider

import (
	"context"
	"crypto/tls"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	changecontrolres "github.com/ioplane/terraform-provider-cvp/internal/resources/change_control"
	inputsres "github.com/ioplane/terraform-provider-cvp/internal/resources/studio_inputs"
	workspaceres "github.com/ioplane/terraform-provider-cvp/internal/resources/workspace"
)

type cvpProvider struct{}

func New() provider.Provider {
	return &cvpProvider{}
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

// ClientBundle propagated to each resource's Configure hook. Wraps
// gRPC conn + auth-method for chargen.v1, studio.v1, workspace.v1, changecontrol.v1.
type ClientBundle struct {
	Conn *grpc.ClientConn
	Auth string // session / cert / bearer
}

func (p *cvpProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cvp"
	resp.Version = "0.1.0-prototype"
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

	creds, diags := buildTLSCreds(&cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn, err := grpc.DialContext(ctx,
		cfg.Endpoint.ValueString(),
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	)
	if err != nil {
		resp.Diagnostics.AddError("gRPC dial failed", err.Error())
		return
	}

	bundle := &ClientBundle{Conn: conn, Auth: cfg.AuthMethod.ValueString()}
	resp.DataSourceData = bundle
	resp.ResourceData = bundle
}

func buildTLSCreds(cfg *providerModel) (credentials.TransportCredentials, diag.Diagnostics) {
	// TODO(P1): implement cert/session/bearer TLS setup.
	// Placeholder — accepts any cert (insecure) for prototype only.
	if cfg.InsecureTLS.ValueBool() {
		return credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}), nil
	}
	return credentials.NewTLS(&tls.Config{}), nil
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
