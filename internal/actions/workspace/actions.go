// Package workspace implements the CVP workspace workflow verbs as Terraform
// actions (ADR 0006). Each action is a one-shot imperative call —
// WorkspaceConfigService.Set with a Request enum and a minted request_id — as
// opposed to the declarative cvp_workspace resource. The verb set is exactly the
// arista.workspace.v1 Request enum: build, cancel_build, submit (+force),
// abandon, rollback, rebase. approve/start are change-control verbs, not here.
package workspace

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ioplane/terraform-provider-cvp/internal/client/cvp"
)

var (
	_ action.Action              = &verbAction{}
	_ action.ActionWithConfigure = &verbAction{}
)

// verbAction exposes one workspace workflow verb as a Terraform action. All
// verbs share this type; the factory functions below configure each one.
type verbAction struct {
	typeSuffix  string
	verb        cvp.WorkspaceVerb
	description string
	allowForce  bool // submit only: force=true escalates to REQUEST_SUBMIT_FORCE
	client      *cvp.Client
}

// NewBuildAction returns the cvp_workspace_build action.
func NewBuildAction() action.Action {
	return &verbAction{
		typeSuffix:  "workspace_build",
		verb:        cvp.VerbBuild,
		description: "Start a build of the workspace (arista.workspace.v1 REQUEST_START_BUILD).",
	}
}

// NewCancelBuildAction returns the cvp_workspace_cancel_build action.
func NewCancelBuildAction() action.Action {
	return &verbAction{
		typeSuffix:  "workspace_cancel_build",
		verb:        cvp.VerbCancelBuild,
		description: "Cancel the in-flight build of the workspace (REQUEST_CANCEL_BUILD).",
	}
}

// NewSubmitAction returns the cvp_workspace_submit action.
func NewSubmitAction() action.Action {
	return &verbAction{
		typeSuffix: "workspace_submit",
		verb:       cvp.VerbSubmit,
		description: "Submit the workspace, creating change controls (REQUEST_SUBMIT; " +
			"set force=true for REQUEST_SUBMIT_FORCE).",
		allowForce: true,
	}
}

// NewAbandonAction returns the cvp_workspace_abandon action.
func NewAbandonAction() action.Action {
	return &verbAction{
		typeSuffix:  "workspace_abandon",
		verb:        cvp.VerbAbandon,
		description: "Abandon the workspace, discarding its pending changes (REQUEST_ABANDON).",
	}
}

// NewRollbackAction returns the cvp_workspace_rollback action.
func NewRollbackAction() action.Action {
	return &verbAction{
		typeSuffix:  "workspace_rollback",
		verb:        cvp.VerbRollback,
		description: "Roll back the workspace (REQUEST_ROLLBACK).",
	}
}

// NewRebaseAction returns the cvp_workspace_rebase action.
func NewRebaseAction() action.Action {
	return &verbAction{
		typeSuffix:  "workspace_rebase",
		verb:        cvp.VerbRebase,
		description: "Rebase the workspace onto the latest mainline (REQUEST_REBASE).",
	}
}

func (a *verbAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + a.typeSuffix
}

func (a *verbAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	attrs := map[string]schema.Attribute{
		"workspace_id": schema.StringAttribute{
			MarkdownDescription: "The `workspace_id` of the target workspace.",
			Required:            true,
		},
	}
	if a.allowForce {
		attrs["force"] = schema.BoolAttribute{
			MarkdownDescription: "Submit even if the workspace has build warnings " +
				"(REQUEST_SUBMIT_FORCE). Defaults to false.",
			Optional: true,
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: a.description,
		Attributes:          attrs,
	}
}

func (a *verbAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*cvp.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			fmt.Sprintf("Expected *cvp.Client for the %s action; this is a provider bug.", a.typeSuffix))
		return
	}
	a.client = client
}

func (a *verbAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var workspaceID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("workspace_id"), &workspaceID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	verb := a.verb
	if a.allowForce {
		var force types.Bool
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("force"), &force)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if force.ValueBool() {
			verb = cvp.VerbSubmitForce
		}
	}

	id := workspaceID.ValueString()
	sendProgress(resp, fmt.Sprintf("cvp: %s workspace %s", verb, id))

	requestID, err := a.client.SubmitWorkspaceVerb(ctx, id, verb)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Workspace %s failed", verb), err.Error())
		return
	}
	sendProgress(resp, fmt.Sprintf("cvp: %s accepted (request_id=%s)", verb, requestID))
}

func sendProgress(resp *action.InvokeResponse, msg string) {
	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: msg})
	}
}
