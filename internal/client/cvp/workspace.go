package cvp

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	sdk "github.com/ioplane/cvp-sdk/go/cvp"
	workspacev1 "github.com/ioplane/cvp-sdk/go/gen/arista/workspace.v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrWorkspaceNotFound = errors.New("cvp: workspace not found")

type WorkspaceInput = sdk.WorkspaceInput

// WorkspaceState is the read-back view of a workspace from WorkspaceService.
// Timestamps are RFC3339 strings ("" when unset).
type WorkspaceState struct {
	ID                         string
	DisplayName                string
	Description                string
	ExcludeNetworkProvisioning bool
	State                      string
	NeedsBuild                 bool
	LastBuildID                string
	CreatedAt                  string
	CreatedBy                  string
	LastModifiedAt             string
	LastModifiedBy             string
	CcIDs                      []string
}

type WorkspaceVerb = sdk.WorkspaceVerb

const (
	VerbBuild       = sdk.VerbBuild
	VerbCancelBuild = sdk.VerbCancelBuild
	VerbSubmit      = sdk.VerbSubmit
	VerbSubmitForce = sdk.VerbSubmitForce
	VerbAbandon     = sdk.VerbAbandon
	VerbRollback    = sdk.VerbRollback
	VerbRebase      = sdk.VerbRebase
)

func (c *Client) SetWorkspace(ctx context.Context, in WorkspaceInput) error {
	return c.sdk.Workspaces().Set(ctx, in)
}

func (c *Client) GetWorkspace(ctx context.Context, id string) (*WorkspaceState, error) {
	state, err := c.sdk.Workspaces().Get(ctx, id)
	if errors.Is(err, sdk.ErrNotFound) {
		return nil, ErrWorkspaceNotFound
	}
	if err != nil {
		return nil, err
	}
	return toWorkspaceState(state), nil
}

// SubmitWorkspaceVerb creates one correlation ID and preserves it on every RPC error.
func (c *Client) SubmitWorkspaceVerb(ctx context.Context, id string, verb WorkspaceVerb) (string, error) {
	idRequest, err := newRequestID()
	if err != nil {
		return "", err
	}
	op, err := c.sdk.Workspaces().Submit(ctx, id, verb, idRequest)
	if err != nil {
		return op.RequestID, fmt.Errorf("cvp: workspace %q %s request_id=%s outcome=%s: %w", id, verb, op.RequestID, op.Outcome, err)
	}
	return op.RequestID, nil
}

// DeleteWorkspace waits for confirmed abandon before deleting a non-terminal workspace.
func (c *Client) DeleteWorkspace(ctx context.Context, id string) error {
	requestID, err := newRequestID()
	if err != nil {
		return err
	}
	if err := c.sdk.Workspaces().AbandonAndDelete(ctx, id, requestID); err != nil {
		return fmt.Errorf("cvp: delete workspace %q request_id=%s: %w", id, requestID, err)
	}
	return nil
}

func toWorkspaceState(w *workspacev1.Workspace) *WorkspaceState {
	if w == nil {
		return &WorkspaceState{}
	}
	return &WorkspaceState{
		ID:                         w.GetKey().GetWorkspaceId().GetValue(),
		DisplayName:                w.GetDisplayName().GetValue(),
		Description:                w.GetDescription().GetValue(),
		ExcludeNetworkProvisioning: w.GetExcludeNetworkProvisioning().GetValue(),
		State:                      w.GetState().String(),
		NeedsBuild:                 w.GetNeedsBuild().GetValue(),
		LastBuildID:                w.GetLastBuildId().GetValue(),
		CreatedAt:                  formatTimestamp(w.GetCreatedAt()),
		CreatedBy:                  w.GetCreatedBy().GetValue(),
		LastModifiedAt:             formatTimestamp(w.GetLastModifiedAt()),
		LastModifiedBy:             w.GetLastModifiedBy().GetValue(),
		CcIDs:                      w.GetCcIds().GetValues(),
	}
}

func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil || !ts.IsValid() {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339)
}

// NewWorkspaceID mints a random workspace_id when the practitioner does not
// supply one. CVP accepts any string as the key; a UUIDv4 keeps it unique.
func NewWorkspaceID() (string, error) { return newRequestID() }

// newRequestID mints a random UUIDv4 for a workflow request. crypto/rand keeps
// it collision-safe without adding a UUID dependency.
func newRequestID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("cvp: generate request id: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
