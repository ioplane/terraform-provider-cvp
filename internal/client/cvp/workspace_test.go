package cvp

import (
	"regexp"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	workspacev1 "github.com/ioplane/cvp-sdk/go/gen/arista/workspace.v1"
	fmp "github.com/ioplane/cvp-sdk/go/gen/fmp"
)

func TestToWorkspaceState(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
	w := &workspacev1.Workspace{
		Key:                        &workspacev1.WorkspaceKey{WorkspaceId: wrapperspb.String("ws-1")},
		DisplayName:                wrapperspb.String("demo"),
		Description:                wrapperspb.String("a description"),
		ExcludeNetworkProvisioning: wrapperspb.Bool(true),
		State:                      workspacev1.WorkspaceState_WORKSPACE_STATE_PENDING,
		NeedsBuild:                 wrapperspb.Bool(true),
		LastBuildId:                wrapperspb.String("build-9"),
		CreatedAt:                  timestamppb.New(created),
		CreatedBy:                  wrapperspb.String("cvpadmin"),
		CcIds:                      &fmp.RepeatedString{Values: []string{"cc-1", "cc-2"}},
	}

	got := toWorkspaceState(w)

	if got.ID != "ws-1" || got.DisplayName != "demo" || got.Description != "a description" {
		t.Fatalf("scalar mapping wrong: %+v", got)
	}
	if !got.ExcludeNetworkProvisioning || !got.NeedsBuild {
		t.Fatalf("bool mapping wrong: %+v", got)
	}
	if got.State != "WORKSPACE_STATE_PENDING" {
		t.Fatalf("state = %q, want WORKSPACE_STATE_PENDING", got.State)
	}
	if got.LastBuildID != "build-9" || got.CreatedBy != "cvpadmin" {
		t.Fatalf("id/audit mapping wrong: %+v", got)
	}
	if got.CreatedAt != "2026-07-28T10:00:00Z" {
		t.Fatalf("createdAt = %q, want RFC3339 UTC", got.CreatedAt)
	}
	if len(got.CcIDs) != 2 || got.CcIDs[0] != "cc-1" {
		t.Fatalf("ccIds mapping wrong: %+v", got.CcIDs)
	}
}

func TestToWorkspaceState_Nil(t *testing.T) {
	t.Parallel()
	got := toWorkspaceState(nil)
	if got == nil {
		t.Fatal("toWorkspaceState(nil) returned nil; want zero-value state")
	}
	if got.ID != "" || got.CreatedAt != "" || len(got.CcIDs) != 0 {
		t.Fatalf("nil workspace should map to zero value, got %+v", got)
	}
}

func TestNewRequestID(t *testing.T) {
	t.Parallel()
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	seen := make(map[string]struct{})
	for range 1000 {
		id, err := newRequestID()
		if err != nil {
			t.Fatalf("newRequestID: %v", err)
		}
		if !re.MatchString(id) {
			t.Fatalf("newRequestID produced non-UUIDv4 %q", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("newRequestID collision on %q", id)
		}
		seen[id] = struct{}{}
	}
}
