package functions

import (
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func obj(t *testing.T, kv map[string]attr.Value) types.Object {
	t.Helper()
	types_ := make(map[string]attr.Type, len(kv))
	for k, v := range kv {
		types_[k] = v.Type(t.Context())
	}
	o, diags := types.ObjectValue(types_, kv)
	if diags.HasError() {
		t.Fatalf("build object: %v", diags)
	}
	return o
}

func TestFormatSegment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   attr.Value
		want string
	}{
		{"group member", types.StringValue("vrfDescription"), "vrfDescription"},
		{"string key", obj(t, map[string]attr.Value{"vrfName": types.StringValue("RED-VRF")}), "[vrfName=RED-VRF]"},
		{"int key", obj(t, map[string]attr.Value{"vlanId": types.NumberValue(big.NewFloat(100))}), "[vlanId=100]"},
		{"bool key", obj(t, map[string]attr.Value{"enabled": types.BoolValue(true)}), "[enabled=true]"},
		{"empty string segment", types.StringValue(""), ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := formatSegment(tc.in)
			if err != nil {
				t.Fatalf("formatSegment(%s) unexpected error: %v", tc.name, err)
			}
			if got != tc.want {
				t.Errorf("formatSegment(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestFormatSegment_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   attr.Value
	}{
		{"multi-key object", obj(t, map[string]attr.Value{"a": types.StringValue("1"), "b": types.StringValue("2")})},
		{"empty object", obj(t, map[string]attr.Value{})},
		{"bare number segment", types.NumberValue(big.NewFloat(5))},
		{"bare bool segment", types.BoolValue(true)},
		{"null string", types.StringNull()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := formatSegment(tc.in); err == nil {
				t.Errorf("formatSegment(%s) = nil error, want error", tc.name)
			}
		})
	}
}

func TestScalarString_UnsupportedValue(t *testing.T) {
	t.Parallel()
	// A list value is not a valid keyed-collection scalar.
	l, diags := types.ListValue(types.StringType, []attr.Value{types.StringValue("x")})
	if diags.HasError() {
		t.Fatalf("build list: %v", diags)
	}
	if _, err := scalarString(l); err == nil {
		t.Error("scalarString(list) = nil error, want error")
	}
}

func TestFormatNumber(t *testing.T) {
	t.Parallel()
	cases := map[string]*big.Float{
		"100":   big.NewFloat(100),
		"0":     big.NewFloat(0),
		"100.5": big.NewFloat(100.5),
		"-7":    big.NewFloat(-7),
	}
	for want, f := range cases {
		if got := formatNumber(f); got != want {
			t.Errorf("formatNumber(%v) = %q, want %q", f, got, want)
		}
	}
	if got := formatNumber(nil); got != "0" {
		t.Errorf("formatNumber(nil) = %q, want 0", got)
	}
}
