// Package studio_inputs implements the `cvp_studio_inputs` resource — a
// declarative inputs value at a Studio path inside a workspace.
//
// Backs arista.studio.v1.InputsConfigService: Create/Update via Set, Read via
// the config GetOne (which round-trips exactly what was written), Delete via the
// config Delete. CVP validates the value against the studio schema on Set — the
// provider passes inputs_json through and surfaces any schema error as a
// diagnostic (design.md: "the build catches those errors").
//
// Key: {studio_id, workspace_id, path[]}. An empty path targets the studio root.
package studio_inputs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ioplane/terraform-provider-cvp/internal/client/cvp"
)

var (
	_ resource.Resource                   = &studioInputsResource{}
	_ resource.ResourceWithConfigure      = &studioInputsResource{}
	_ resource.ResourceWithImportState    = &studioInputsResource{}
	_ resource.ResourceWithValidateConfig = &studioInputsResource{}
)

type studioInputsResource struct {
	client *cvp.Client
}

func NewResource() resource.Resource { return &studioInputsResource{} }

type studioInputsModel struct {
	ID          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	StudioID    types.String `tfsdk:"studio_id"`
	Path        types.List   `tfsdk:"path"`
	InputsJSON  types.String `tfsdk:"inputs_json"`
}

func (r *studioInputsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_studio_inputs"
}

func (r *studioInputsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A declarative inputs value at a Studio path inside a workspace " +
			"(`arista.studio.v1.InputsConfigService`). The key `{studio_id, workspace_id, path}` " +
			"is immutable; changing any part replaces the resource. CVP validates `inputs_json` " +
			"against the studio schema on write.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite id `{workspace_id}/{studio_id}/{path...}`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "`workspace_id` from a `cvp_workspace`.",
				Required:            true,
				PlanModifiers:       replace,
			},
			"studio_id": schema.StringAttribute{
				MarkdownDescription: "Studio id (e.g. `studio-l3ls`). The studio is referenced from mainline.",
				Required:            true,
				PlanModifiers:       replace,
			},
			"path": schema.ListAttribute{
				MarkdownDescription: "Ordered path segments to the input. Empty targets the studio root. " +
					"Bracket notation is supported for keyed collections and resolvers.",
				ElementType:   types.StringType,
				Required:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"inputs_json": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded value at `path` (use `jsonencode(...)`). CVP validates it " +
					"against the studio schema on write; the provider does not.",
				Required: true,
			},
		},
	}
}

func (r *studioInputsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*cvp.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data",
			"Expected *cvp.Client for the cvp_studio_inputs resource; this is a provider bug.")
		return
	}
	r.client = client
}

// ValidateConfig catches invalid JSON at pre-plan.
func (r *studioInputsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg studioInputsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !cfg.InputsJSON.IsNull() && !cfg.InputsJSON.IsUnknown() {
		var v any
		if err := json.Unmarshal([]byte(cfg.InputsJSON.ValueString()), &v); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("inputs_json"), "inputs_json is not valid JSON", err.Error())
		}
	}
}

func (r *studioInputsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan studioInputsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, diags := plan.toInput(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.SetInputs(ctx, in); err != nil {
		resp.Diagnostics.AddError("Cannot set studio inputs", err.Error())
		return
	}
	plan.ID = types.StringValue(buildID(in.WorkspaceID, in.StudioID, in.Path))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *studioInputsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state studioInputsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pathSegs, diags := listToStrings(ctx, state.Path)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	st, err := r.client.GetInputs(ctx, state.StudioID.ValueString(), state.WorkspaceID.ValueString(), pathSegs)
	if err != nil {
		if errors.Is(err, cvp.ErrInputsNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Cannot read studio inputs", err.Error())
		return
	}
	state.InputsJSON = types.StringValue(st.InputsJSON)
	state.ID = types.StringValue(buildID(st.WorkspaceID, st.StudioID, st.Path))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *studioInputsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan studioInputsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in, diags := plan.toInput(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.SetInputs(ctx, in); err != nil {
		resp.Diagnostics.AddError("Cannot update studio inputs", err.Error())
		return
	}
	plan.ID = types.StringValue(buildID(in.WorkspaceID, in.StudioID, in.Path))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *studioInputsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state studioInputsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pathSegs, diags := listToStrings(ctx, state.Path)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteInputs(ctx, state.StudioID.ValueString(), state.WorkspaceID.ValueString(), pathSegs); err != nil {
		resp.Diagnostics.AddError("Cannot delete studio inputs", err.Error())
	}
}

// ImportState parses the composite id `{workspace_id}/{studio_id}/{path...}`.
func (r *studioInputsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import id",
			fmt.Sprintf("Expected `{workspace_id}/{studio_id}/{path...}`, got %q.", req.ID))
		return
	}
	segs := parts[2:]
	pathVal, diags := types.ListValueFrom(ctx, types.StringType, segs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("studio_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("path"), pathVal)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (m studioInputsModel) toInput(ctx context.Context) (cvp.InputsInput, diag.Diagnostics) {
	pathSegs, diags := listToStrings(ctx, m.Path)
	return cvp.InputsInput{
		StudioID:    m.StudioID.ValueString(),
		WorkspaceID: m.WorkspaceID.ValueString(),
		Path:        pathSegs,
		InputsJSON:  m.InputsJSON.ValueString(),
	}, diags
}

func listToStrings(ctx context.Context, l types.List) ([]string, diag.Diagnostics) {
	segs := []string{}
	if l.IsNull() || l.IsUnknown() {
		return segs, nil
	}
	diags := l.ElementsAs(ctx, &segs, false)
	return segs, diags
}

// buildID renders the composite id. Path segments are bracket-notation and do
// not contain '/', so a '/' join is unambiguous for import.
func buildID(workspaceID, studioID string, path []string) string {
	return strings.Join(append([]string{workspaceID, studioID}, path...), "/")
}
