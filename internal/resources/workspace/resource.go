// Package workspace implements the `cvp_workspace` resource.
//
// Backs `arista.workspace.v1.WorkspaceConfigService`. Один resource =
// один workspace в CVP; создание = CreateWorkspace RPC; удаление =
// AbandonWorkspace. Submit/build/approve — отдельные resources в v0.2+.
package workspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &workspaceResource{}

type workspaceResource struct {
	// clients набор для workspace.v1 gRPC
}

func NewResource() resource.Resource { return &workspaceResource{} }

type workspaceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	// Explicit workflow gates — provider-side helpers, not schema fields on CVP wire.
	AutoBuild   types.Bool `tfsdk:"auto_build"`
	AutoSubmit  types.Bool `tfsdk:"auto_submit"`
	AutoApprove types.Bool `tfsdk:"auto_approve"`
}

func (r *workspaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `CVP workspace — atomic mutation scope.
См. docs/notes/2026-07-16-i2-studios-reverse.md, section 7.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "workspace_id (UUID from CVP).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable name; per naming standard `[<squad>] PR#<ID> <text>`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description shown в CVP UI.",
			},
			"auto_build": schema.BoolAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Trigger BuildStatus poll after Set. Default true.",
			},
			"auto_submit": schema.BoolAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Auto-submit после successful build. Default false.",
			},
			"auto_approve": schema.BoolAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Auto-approve после submit — **breaks separation of duties**. " +
					"Set only для test env.",
			},
		},
	}
}

func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO(P1): call workspace.v1.WorkspaceConfigService.Set
	//    key = { workspace_id: "" }  → CVP mints new UUID
	//    display_name, description
	//    request = REQUEST_UNSPECIFIED
	// Returns workspace_id → plan.ID
	//
	// If plan.AutoBuild → subsequent Set(request=REQUEST_START_BUILD)
	// If plan.AutoSubmit → subsequent Set(request=REQUEST_SUBMIT) — separate call.

	resp.Diagnostics.AddError("not implemented",
		"workspace.Create — Aeris gRPC stub pending. See TODO(P1).")
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO(P1): call workspace.v1.WorkspaceService.GetOne
	// key = { workspace_id: state.ID }
	// Detects drift + returns updated display_name/description.

	resp.Diagnostics.AddError("not implemented", "workspace.Read — pending.")
}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO(P1): Set with new display_name/description; workspace_id unchanged.

	resp.Diagnostics.AddError("not implemented", "workspace.Update — pending.")
}

func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO(P1): call workspace.v1.WorkspaceConfigService.Set с request=REQUEST_ABANDON
	// Не удаляет workspace physically (workspaces immutable в CVP); маркирует ABANDONED.

	resp.Diagnostics.AddError("not implemented", "workspace.Delete — pending.")
}

func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
