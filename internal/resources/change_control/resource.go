// Package change_control implements `cvp_change_control` resource —
// bundle of rendered device configs + schedule + start.
package change_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &changeControlResource{}

type changeControlResource struct{}

func NewResource() resource.Resource { return &changeControlResource{} }

type changeControlModel struct {
	ID                 types.String `tfsdk:"id"`
	WorkspaceID        types.String `tfsdk:"workspace_id"`
	Description        types.String `tfsdk:"description"`
	AutoApprove        types.Bool   `tfsdk:"auto_approve"`
	WaitForExecution   types.Bool   `tfsdk:"wait_for_execution"`
	StageExecution     types.String `tfsdk:"stage_execution"` // parallel|rolling|sequential
	StopOnFailure      types.Bool   `tfsdk:"stop_on_failure"`
	RollbackOnFailure  types.Bool   `tfsdk:"rollback_on_failure"`
	ScheduleAt         types.String `tfsdk:"schedule_at"`
	Status             types.String `tfsdk:"status"`
	ExecutionError     types.String `tfsdk:"execution_error"`
}

func (r *changeControlResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_change_control"
}

func (r *changeControlResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `ChangeControl — bundle of rendered device configs + schedule + approval + start.
См. docs/integration-drafts/um-docs/07-architecture/cvp/06-workspace-changecontrol.md.`,
		Attributes: map[string]schema.Attribute{
			"id":                    schema.StringAttribute{Computed: true},
			"workspace_id":          schema.StringAttribute{Required: true},
			"description":           schema.StringAttribute{Optional: true},
			"auto_approve":          schema.BoolAttribute{Optional: true, Computed: true},
			"wait_for_execution":    schema.BoolAttribute{Optional: true, Computed: true},
			"stage_execution":       schema.StringAttribute{Optional: true, Computed: true},
			"stop_on_failure":       schema.BoolAttribute{Optional: true, Computed: true},
			"rollback_on_failure":   schema.BoolAttribute{Optional: true, Computed: true},
			"schedule_at":           schema.StringAttribute{Optional: true},
			"status":                schema.StringAttribute{Computed: true},
			"execution_error":       schema.StringAttribute{Computed: true},
		},
	}
}

func (r *changeControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TODO(P1):
	//   1. changecontrol.v1.ChangeControlConfigService.Set — create CC record с workspace_id.
	//   2. If auto_approve → changecontrol.v1.ApproveConfigService.Set — approve.
	//   3. changecontrol.v1.ChangeControlConfigService.Set с start=true.
	//   4. If wait_for_execution → poll ChangeControlService.GetOne until status terminal.
	resp.Diagnostics.AddError("not implemented", "change_control.Create — pending.")
}

func (r *changeControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError("not implemented", "change_control.Read — pending.")
}

func (r *changeControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// ChangeControl мостdirty — большинство fields immutable после Start. TF marks as ForceNew.
	resp.Diagnostics.AddError("not implemented", "change_control.Update — pending.")
}

func (r *changeControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// В CVP CC persists indefinitely (audit). TF state delete = no-op with warning.
	resp.Diagnostics.AddWarning("no-op",
		"ChangeControl records persist в CVP для audit purposes. Terraform state deletion не удаляет запись.")
}
