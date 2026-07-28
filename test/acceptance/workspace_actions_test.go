//go:build acceptance

package acceptance

import (
	"context"
	"testing"
	"time"

	workspacev1 "github.com/ioplane/terraform-provider-cvp/internal/pb/arista/workspace.v1"

	"github.com/ioplane/terraform-provider-cvp/internal/client/cvp"
)

// TestAccWorkspaceVerbs_buildAbandon exercises the imperative verb path that the
// cvp_workspace_* actions invoke (SubmitWorkspaceVerb): build echoes the minted
// request_id back as last_build_id, and abandon drives the workspace to the
// ABANDONED terminal state. Evidence for ADR 0006.
func TestAccWorkspaceVerbs_buildAbandon(t *testing.T) {
	preCheck(t)
	client := accClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	id, err := cvp.NewWorkspaceID()
	if err != nil {
		t.Fatalf("mint workspace_id: %v", err)
	}
	if err = client.SetWorkspace(ctx, cvp.WorkspaceInput{ID: id, DisplayName: "tf-acc-ws-verbs"}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() {
		if cerr := client.DeleteWorkspace(context.Background(), id); cerr != nil {
			t.Errorf("cleanup workspace %s: %v", id, cerr)
		}
	})

	// build — request_id is echoed as last_build_id.
	buildID, err := client.SubmitWorkspaceVerb(ctx, id, cvp.VerbBuild)
	if err != nil {
		t.Fatalf("build verb: %v", err)
	}
	if got := pollLastBuildID(ctx, t, client, id); got != buildID {
		t.Fatalf("last_build_id = %q, want the build request_id %q", got, buildID)
	}

	// abandon — terminal ABANDONED state.
	if _, err := client.SubmitWorkspaceVerb(ctx, id, cvp.VerbAbandon); err != nil {
		t.Fatalf("abandon verb: %v", err)
	}
	want := workspacev1.WorkspaceState_WORKSPACE_STATE_ABANDONED.String()
	if got := pollState(ctx, t, client, id, want); got != want {
		t.Fatalf("state = %q, want %q after abandon", got, want)
	}
}

func pollLastBuildID(ctx context.Context, t *testing.T, client *cvp.Client, id string) string {
	t.Helper()
	var last string
	for range 20 {
		st, err := client.GetWorkspace(ctx, id)
		if err != nil {
			t.Fatalf("get workspace: %v", err)
		}
		if last = st.LastBuildID; last != "" {
			return last
		}
		time.Sleep(1 * time.Second)
	}
	return last
}

func pollState(ctx context.Context, t *testing.T, client *cvp.Client, id, want string) string {
	t.Helper()
	var state string
	for range 20 {
		st, err := client.GetWorkspace(ctx, id)
		if err != nil {
			t.Fatalf("get workspace: %v", err)
		}
		if state = st.State; state == want {
			return state
		}
		time.Sleep(1 * time.Second)
	}
	return state
}
