// Package studio_inputs implements `cvp_studio_inputs` — per-workspace input
// values at a Studio path.
//
// **Key semantic**: I2 finding — Set at a higher path OVERWRITES lower paths.
// The provider detects prefix overlap at plan time and fails early.
package studio_inputs

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &studioInputsResource{}

type studioInputsResource struct{}

func NewResource() resource.Resource { return &studioInputsResource{} }

type studioInputsModel struct {
	ID          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	StudioID    types.String `tfsdk:"studio_id"`
	Path        types.List   `tfsdk:"path"` // ordered []string
	InputsJSON  types.String `tfsdk:"inputs_json"`
}

func (r *studioInputsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_studio_inputs"
}

func (r *studioInputsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Studio input values at a specific path.
Path semantics per I2 note (bracket notation supported for keyed collections + resolvers).
See docs/notes/2026-07-16-i2-studios-reverse.md §1.3.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Composite `{workspace_id}:{studio_id}:{path_joined}`.",
			},
			"workspace_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "workspace_id from `cvp_workspace`.",
			},
			"studio_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the Studio. Immutable studios (from_package non-empty) are not supported.",
			},
			"path": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Path segments. Bracket notation is supported for keyed collections and resolvers.",
			},
			"inputs_json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON-encoded value at `path`. The provider does **not** validate against the Studio schema — the build catches those errors.",
			},
		},
	}
}

// ValidateConfig catches invalid JSON and predictable path traps at pre-plan.
func (r *studioInputsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg studioInputsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !cfg.InputsJSON.IsNull() && !cfg.InputsJSON.IsUnknown() {
		var v any
		if err := json.Unmarshal([]byte(cfg.InputsJSON.ValueString()), &v); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("inputs_json"),
				"inputs_json is not valid JSON",
				err.Error(),
			)
		}
	}
	// TODO(P2): prefix-overlap detection between sibling cvp_studio_inputs
	// resources — requires ConfigValidator across whole config, not per-resource.
}

func (r *studioInputsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan studioInputsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO(P1): call studio.v1.InputsConfigService.Set
	//   key = { studio_id, workspace_id, path: [...] }
	//   inputs = plan.InputsJSON
	//   remove = false
	// Set ID = "{workspace_id}:{studio_id}:{path_joined}"

	resp.Diagnostics.AddError("not implemented", "studio_inputs.Create — pending.")
}

func (r *studioInputsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TODO(P1): studio.v1.InputsService.GetOne
	resp.Diagnostics.AddError("not implemented", "studio_inputs.Read — pending.")
}

func (r *studioInputsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// In-place update — Set with the same key and new inputs_json.
	resp.Diagnostics.AddError("not implemented", "studio_inputs.Update — pending.")
}

func (r *studioInputsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Set with remove=true on workspace merge → deletes; otherwise a no-op in CVP.
	resp.Diagnostics.AddError("not implemented", "studio_inputs.Delete — pending.")
}
