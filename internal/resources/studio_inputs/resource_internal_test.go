package studio_inputs

import (
	"strings"
	"testing"
)

func TestBuildID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		ws, studio string
		path       []string
		want       string
	}{
		{"ws1", "studio-l3ls", nil, "ws1/studio-l3ls"},
		{"ws1", "studio-l3ls", []string{}, "ws1/studio-l3ls"},
		{"ws1", "studio-date-time", []string{"timezoneResolver"}, "ws1/studio-date-time/timezoneResolver"},
		{"ws1", "s", []string{"tenants", "[name=default]", "vrfs"}, "ws1/s/tenants/[name=default]/vrfs"},
	}
	for _, c := range cases {
		if got := buildID(c.ws, c.studio, c.path); got != c.want {
			t.Errorf("buildID(%q,%q,%v) = %q, want %q", c.ws, c.studio, c.path, got, c.want)
		}
	}
}

// TestBuildID_ImportRoundTrip checks the id renders and splits back to the same
// key parts (the ImportState parse is the inverse of buildID).
func TestBuildID_ImportRoundTrip(t *testing.T) {
	t.Parallel()
	ws, studio, path := "ws-42", "studio-date-time", []string{"tenants", "[name=default]"}
	id := buildID(ws, studio, path)

	parts := strings.Split(id, "/")
	if parts[0] != ws || parts[1] != studio {
		t.Fatalf("parsed ws/studio = %q/%q, want %q/%q", parts[0], parts[1], ws, studio)
	}
	got := parts[2:]
	if len(got) != len(path) {
		t.Fatalf("parsed path len %d, want %d", len(got), len(path))
	}
	for i := range path {
		if got[i] != path[i] {
			t.Errorf("path[%d] = %q, want %q", i, got[i], path[i])
		}
	}
}
