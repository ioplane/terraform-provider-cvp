// Package functions holds the provider-defined functions (Terraform 1.8+).
package functions

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ function.Function = &studioPathFunction{}

// NewStudioPathFunction is the factory wired into provider.Functions().
func NewStudioPathFunction() function.Function { return &studioPathFunction{} }

type studioPathFunction struct{}

var (
	errSegmentType  = errors.New("a segment must be a string (group member) or a single-key object (keyed collection)")
	errKeyedSegment = errors.New("a keyed-collection segment must be an object with exactly one key")
	errValueType    = errors.New("a keyed-collection value must be a string, number, or bool")
	errUnknownValue = errors.New("segment values must be known (not null or unknown)")
)

func (f *studioPathFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "studio_path"
}

func (f *studioPathFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Build a Studio input path from segments.",
		MarkdownDescription: "Assembles a `cvp_studio_inputs` `path` (a `list(string)`) from ordered segments. " +
			"A string segment is emitted verbatim (a group member or resolver id); a single-key object " +
			"segment becomes bracket key-notation, e.g. `{ vrfName = \"RED-VRF\" }` -> `\"[vrfName=RED-VRF]\"`. " +
			"With no segments it returns `[]` (the studio root). The function is generic and offline — it does " +
			"not validate against any studio schema (CVP does that on write).",
		VariadicParameter: function.DynamicParameter{
			Name:                "segment",
			MarkdownDescription: "A path segment: a string, or a single-key object `{ key = value }`.",
		},
		Return: function.ListReturn{ElementType: types.StringType},
	}
}

func (f *studioPathFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var segments []types.Dynamic
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &segments))
	if resp.Error != nil {
		return
	}

	out := make([]string, 0, len(segments))
	for i, seg := range segments {
		if seg.IsNull() || seg.IsUnknown() || seg.IsUnderlyingValueNull() || seg.IsUnderlyingValueUnknown() {
			resp.Error = function.ConcatFuncErrors(resp.Error,
				function.NewArgumentFuncError(int64(i), errUnknownValue.Error()))
			return
		}
		s, err := formatSegment(seg.UnderlyingValue())
		if err != nil {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(int64(i), err.Error()))
			return
		}
		out = append(out, s)
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, out))
}

// formatSegment renders one path segment from its underlying value. Pure and
// framework-independent so it is unit-tested directly.
func formatSegment(v attr.Value) (string, error) {
	switch val := v.(type) {
	case basetypes.StringValue:
		if val.IsNull() || val.IsUnknown() {
			return "", errUnknownValue
		}
		return val.ValueString(), nil
	case basetypes.ObjectValue:
		attrs := val.Attributes()
		if len(attrs) != 1 {
			return "", fmt.Errorf("%w (got %d)", errKeyedSegment, len(attrs))
		}
		for key, av := range attrs {
			s, err := scalarString(av)
			if err != nil {
				return "", err
			}
			return "[" + key + "=" + s + "]", nil
		}
		return "", errKeyedSegment // unreachable (len==1)
	default:
		return "", errSegmentType
	}
}

// scalarString stringifies a keyed-collection value (string/number/bool).
func scalarString(v attr.Value) (string, error) {
	switch val := v.(type) {
	case basetypes.StringValue:
		if val.IsNull() || val.IsUnknown() {
			return "", errUnknownValue
		}
		return val.ValueString(), nil
	case basetypes.NumberValue:
		if val.IsNull() || val.IsUnknown() {
			return "", errUnknownValue
		}
		return formatNumber(val.ValueBigFloat()), nil
	case basetypes.BoolValue:
		if val.IsNull() || val.IsUnknown() {
			return "", errUnknownValue
		}
		return strconv.FormatBool(val.ValueBool()), nil
	default:
		return "", errValueType
	}
}

// formatNumber renders a number with the minimum digits (100 -> "100").
func formatNumber(f *big.Float) string {
	if f == nil {
		return "0"
	}
	return f.Text('f', -1)
}
