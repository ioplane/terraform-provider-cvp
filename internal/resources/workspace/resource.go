// Package workspace implements the `cvp_workspace` resource.
//
// A workspace is CVP's atomic change scope. This resource manages the workspace
// as declarative state (display name, description, network-provisioning flag)
// over arista.workspace.v1: Create/Update via WorkspaceConfigService.Set, Read
// via WorkspaceService.GetOne, Delete via abandon-then-config-delete. The
// imperative workflow verbs (build, submit, abandon, ...) are Terraform Actions,
// not attributes — see ADR 0006 and internal/actions/workspace.
package workspace

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ioplane/terraform-provider-cvp/internal/client/cvp"
)

var (
	_ resource.Resource                = &workspaceResource{}
	_ resource.ResourceWithConfigure   = &workspaceResource{}
	_ resource.ResourceWithImportState = &workspaceResource{}
)

type workspaceResource struct {
	client *cvp.Client
}

func NewResource() resource.Resource { return &workspaceResource{} }

type workspaceModel struct {
	WorkspaceID                types.String `tfsdk:"workspace_id"`
	DisplayName                types.String `tfsdk:"display_name"`
	Description                types.String `tfsdk:"description"`
	ExcludeNetworkProvisioning types.Bool   `tfsdk:"exclude_network_provisioning"`
	State                      types.String `tfsdk:"state"`
	NeedsBuild                 types.Bool   `tfsdk:"needs_build"`
	LastBuildID                types.String `tfsdk:"last_build_id"`
	CreatedAt                  types.String `tfsdk:"created_at"`
	CreatedBy                  types.String `tfsdk:"created_by"`
	LastModifiedAt             types.String `tfsdk:"last_modified_at"`
	LastModifiedBy             types.String `tfsdk:"last_modified_by"`
	CcIDs                      types.List   `tfsdk:"cc_ids"`
}

func (r *workspaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A CVP workspace — the atomic scope for a set of configuration changes. " +
			"This resource manages the workspace's declarative state; the workflow verbs " +
			"(build, submit, abandon, rebase, rollback) are provider **actions** (ADR 0006).",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "Stable workspace key. Optional — when omitted the provider mints a " +
					"UUIDv4. Changing it replaces the workspace.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Human-readable workspace name shown in the CVP UI.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Free-text description shown in the CVP UI.",
				Optional:            true,
				Computed:            true,
			},
			"exclude_network_provisioning": schema.BoolAttribute{
				MarkdownDescription: "Exclude network-provisioning changes from this workspace's build.",
				Optional:            true,
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Workspace lifecycle state (e.g. `WORKSPACE_STATE_PENDING`, " +
					"`WORKSPACE_STATE_SUBMITTED`, `WORKSPACE_STATE_ABANDONED`).",
				Computed: true,
			},
			"needs_build": schema.BoolAttribute{
				MarkdownDescription: "True when the workspace has changes that have not been built.",
				Computed:            true,
			},
			"last_build_id": schema.StringAttribute{
				MarkdownDescription: "Request id of the most recent build (see the `cvp_workspace_build` action).",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp (RFC3339).",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "User that created the workspace.",
				Computed:            true,
			},
			"last_modified_at": schema.StringAttribute{
				MarkdownDescription: "Last-modification timestamp (RFC3339).",
				Computed:            true,
			},
			"last_modified_by": schema.StringAttribute{
				MarkdownDescription: "User that last modified the workspace.",
				Computed:            true,
			},
			"cc_ids": schema.ListAttribute{
				MarkdownDescription: "Change-control ids created from this workspace on submit.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (r *workspaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*cvp.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			"Expected *cvp.Client for the cvp_workspace resource; this is a provider bug.")
		return
	}
	r.client = client
}

func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.WorkspaceID.ValueString()
	userSupplied := !plan.WorkspaceID.IsUnknown() && !plan.WorkspaceID.IsNull() && id != ""
	if userSupplied {
		// Brownfield guard: refuse to adopt/overwrite an existing workspace via
		// the upsert Set. A pre-existing id must be brought in with import.
		switch _, err := r.client.GetWorkspace(ctx, id); {
		case err == nil:
			resp.Diagnostics.AddError("Workspace already exists",
				"A workspace with workspace_id "+id+" already exists in CVP. Import it with "+
					"`terraform import cvp_workspace.<name> "+id+"` instead of creating it.")
			return
		case !errors.Is(err, cvp.ErrWorkspaceNotFound):
			resp.Diagnostics.AddError("Cannot check workspace existence", err.Error())
			return
		}
	} else {
		generated, err := cvp.NewWorkspaceID()
		if err != nil {
			resp.Diagnostics.AddError("Cannot generate workspace_id", err.Error())
			return
		}
		id = generated
	}

	if err := r.client.SetWorkspace(ctx, plan.toInput(id)); err != nil {
		resp.Diagnostics.AddError("Cannot create workspace", err.Error())
		return
	}

	st, err := r.client.GetWorkspace(ctx, id)
	if err != nil {
		// The workspace exists in CVP but could not be read back (transient or
		// eventual-consistency). Persist a minimal but fully-known state keyed by
		// the id so Terraform tracks the object for a later refresh or destroy —
		// otherwise it leaks as an untracked CVP workspace.
		partial := &cvp.WorkspaceState{
			ID:                         id,
			DisplayName:                plan.DisplayName.ValueString(),
			Description:                plan.Description.ValueString(),
			ExcludeNetworkProvisioning: plan.ExcludeNetworkProvisioning.ValueBool(),
		}
		resp.Diagnostics.Append(applyState(ctx, partial, &plan)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("Workspace created but read-back failed",
			"The workspace was created and recorded in state (workspace_id "+id+"), but reading "+
				"it back failed; a subsequent plan/apply will reconcile it. Error: "+err.Error())
		return
	}
	resp.Diagnostics.Append(applyState(ctx, st, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	st, err := r.client.GetWorkspace(ctx, state.WorkspaceID.ValueString())
	if err != nil {
		if errors.Is(err, cvp.ErrWorkspaceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Cannot read workspace", err.Error())
		return
	}
	resp.Diagnostics.Append(applyState(ctx, st, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.WorkspaceID.ValueString()
	if err := r.client.SetWorkspace(ctx, plan.toInput(id)); err != nil {
		resp.Diagnostics.AddError("Cannot update workspace", err.Error())
		return
	}

	st, err := r.client.GetWorkspace(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Workspace updated but read-back failed", err.Error())
		return
	}
	resp.Diagnostics.Append(applyState(ctx, st, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteWorkspace(ctx, state.WorkspaceID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Cannot delete workspace", err.Error())
	}
}

func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), req, resp)
}

// toInput projects the plan onto the client's declarative input.
func (m workspaceModel) toInput(id string) cvp.WorkspaceInput {
	return cvp.WorkspaceInput{
		ID:                         id,
		DisplayName:                m.DisplayName.ValueString(),
		Description:                m.Description.ValueString(),
		ExcludeNetworkProvisioning: m.ExcludeNetworkProvisioning.ValueBool(),
	}
}

// applyState copies the server read-back into the model.
func applyState(ctx context.Context, st *cvp.WorkspaceState, m *workspaceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	m.WorkspaceID = types.StringValue(st.ID)
	m.DisplayName = types.StringValue(st.DisplayName)
	m.Description = types.StringValue(st.Description)
	m.ExcludeNetworkProvisioning = types.BoolValue(st.ExcludeNetworkProvisioning)
	m.State = types.StringValue(st.State)
	m.NeedsBuild = types.BoolValue(st.NeedsBuild)
	m.LastBuildID = types.StringValue(st.LastBuildID)
	m.CreatedAt = types.StringValue(st.CreatedAt)
	m.CreatedBy = types.StringValue(st.CreatedBy)
	m.LastModifiedAt = types.StringValue(st.LastModifiedAt)
	m.LastModifiedBy = types.StringValue(st.LastModifiedBy)

	ccIDs, d := types.ListValueFrom(ctx, types.StringType, st.CcIDs)
	diags.Append(d...)
	m.CcIDs = ccIDs
	return diags
}
