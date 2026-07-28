package studio_inputs

import (
	"slices"
	"testing"
)

// TestID_RoundTrip checks buildID/parseID are inverses, including resolver path
// segments that contain '/' (studio.proto bracket notation).
func TestID_RoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		ws, studio string
		path       []string
	}{
		{"ws1", "studio-l3ls", nil},
		{"ws1", "studio-l3ls", []string{}},
		{"ws-42", "studio-date-time", []string{"timezoneResolver"}},
		{"ws1", "s", []string{"tenants", "[name=default]", "vrfs", "[name=RED]"}},
		{"ws1", "s", []string{"devices", "[tags/query=device:leaf1]", "inputs", "hostname"}},
	}
	for _, c := range cases {
		id := buildID(c.ws, c.studio, c.path)
		ws, studio, path, err := parseID(id)
		if err != nil {
			t.Fatalf("parseID(%q): %v", id, err)
		}
		if ws != c.ws || studio != c.studio {
			t.Errorf("id %q -> ws/studio %q/%q, want %q/%q", id, ws, studio, c.ws, c.studio)
		}
		want := c.path
		if want == nil {
			want = []string{}
		}
		if !slices.Equal(path, want) {
			t.Errorf("id %q -> path %v, want %v", id, path, want)
		}
	}
}

func TestParseID_Invalid(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"", "onlyworkspace", "/studio", "ws/"} {
		if _, _, _, err := parseID(id); err == nil {
			t.Errorf("parseID(%q) = nil error, want error", id)
		}
	}
}

func TestPathIsPrefixConflict(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b []string
		want bool
	}{
		{[]string{"A"}, []string{"A", "X"}, true},           // strict prefix
		{[]string{"A", "X"}, []string{"A"}, true},           // strict prefix (reversed)
		{[]string{}, []string{"A"}, true},                   // root vs child
		{[]string{"A"}, []string{"A"}, false},               // equal — not a conflict
		{[]string{"A"}, []string{"B"}, false},               // disjoint
		{[]string{"A", "X"}, []string{"A", "Y"}, false},     // siblings
		{[]string{"A", "X"}, []string{"A", "X", "Z"}, true}, // strict prefix, deeper
	}
	for _, c := range cases {
		if got := pathIsPrefixConflict(c.a, c.b); got != c.want {
			t.Errorf("pathIsPrefixConflict(%v,%v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
