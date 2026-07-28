package cvp

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	workspacev1 "github.com/ioplane/terraform-provider-cvp/internal/pb/arista/workspace.v1"
)

// ErrWorkspaceNotFound is returned by GetWorkspace when CVP reports the
// workspace does not exist (gRPC NotFound). Callers map it to state removal.
var ErrWorkspaceNotFound = errors.New("cvp: workspace not found")

// WorkspaceInput is the declarative desired state of a workspace — the config
// the cvp_workspace resource converges via WorkspaceConfigService.Set.
type WorkspaceInput struct {
	ID                         string
	DisplayName                string
	Description                string
	ExcludeNetworkProvisioning bool
}

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

// WorkspaceVerb is an imperative workspace workflow request (ADR 0006). Each
// verb is one WorkspaceConfigService.Set carrying a Request enum plus a minted
// request_id (CVP rejects a Request without one: "request ID parameter is
// missing").
type WorkspaceVerb string

const (
	VerbBuild       WorkspaceVerb = "build"
	VerbCancelBuild WorkspaceVerb = "cancel_build"
	VerbSubmit      WorkspaceVerb = "submit"
	VerbSubmitForce WorkspaceVerb = "submit_force"
	VerbAbandon     WorkspaceVerb = "abandon"
	VerbRollback    WorkspaceVerb = "rollback"
	VerbRebase      WorkspaceVerb = "rebase"
)

// verbToRequest maps a WorkspaceVerb to its wire Request enum.
var verbToRequest = map[WorkspaceVerb]workspacev1.Request{
	VerbBuild:       workspacev1.Request_REQUEST_START_BUILD,
	VerbCancelBuild: workspacev1.Request_REQUEST_CANCEL_BUILD,
	VerbSubmit:      workspacev1.Request_REQUEST_SUBMIT,
	VerbSubmitForce: workspacev1.Request_REQUEST_SUBMIT_FORCE,
	VerbAbandon:     workspacev1.Request_REQUEST_ABANDON,
	VerbRollback:    workspacev1.Request_REQUEST_ROLLBACK,
	VerbRebase:      workspacev1.Request_REQUEST_REBASE,
}

func workspaceKey(id string) *workspacev1.WorkspaceKey {
	return &workspacev1.WorkspaceKey{WorkspaceId: wrapperspb.String(id)}
}

// SetWorkspace creates or updates a workspace config (the declarative fields).
// It never carries a Request verb — those go through SubmitWorkspaceVerb.
func (c *Client) SetWorkspace(ctx context.Context, in WorkspaceInput) error {
	_, err := c.WorkspaceConfig().Set(ctx, &workspacev1.WorkspaceConfigSetRequest{
		Value: &workspacev1.WorkspaceConfig{
			Key:                        workspaceKey(in.ID),
			DisplayName:                wrapperspb.String(in.DisplayName),
			Description:                wrapperspb.String(in.Description),
			ExcludeNetworkProvisioning: wrapperspb.Bool(in.ExcludeNetworkProvisioning),
		},
	})
	if err != nil {
		return fmt.Errorf("cvp: set workspace %q: %w", in.ID, err)
	}
	return nil
}

// GetWorkspace reads a workspace's full state. A NotFound is translated to
// ErrWorkspaceNotFound so callers can distinguish "gone" from a transport error.
func (c *Client) GetWorkspace(ctx context.Context, id string) (*WorkspaceState, error) {
	resp, err := c.Workspace().GetOne(ctx, &workspacev1.WorkspaceRequest{Key: workspaceKey(id)})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("cvp: get workspace %q: %w", id, err)
	}
	return toWorkspaceState(resp.GetValue()), nil
}

// SubmitWorkspaceVerb issues one imperative workflow verb and returns the minted
// request_id (for build, CVP echoes it back as last_build_id).
func (c *Client) SubmitWorkspaceVerb(ctx context.Context, id string, verb WorkspaceVerb) (string, error) {
	req, ok := verbToRequest[verb]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrUnknownWorkspaceVerb, verb)
	}
	requestID, err := newRequestID()
	if err != nil {
		return "", err
	}
	_, err = c.WorkspaceConfig().Set(ctx, &workspacev1.WorkspaceConfigSetRequest{
		Value: &workspacev1.WorkspaceConfig{
			Key:           workspaceKey(id),
			Request:       req,
			RequestParams: &workspacev1.RequestParams{RequestId: wrapperspb.String(requestID)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("cvp: workspace %q %s: %w", id, verb, err)
	}
	return requestID, nil
}

// DeleteWorkspace tears a workspace down cleanly (ADR 0006, evidence-backed):
// abandon a non-terminal workspace first, then delete the config. Config delete
// alone leaves a PENDING workspace in place; abandon-then-delete yields NotFound.
// A missing workspace is treated as already gone.
func (c *Client) DeleteWorkspace(ctx context.Context, id string) error {
	st, err := c.GetWorkspace(ctx, id)
	switch {
	case errors.Is(err, ErrWorkspaceNotFound):
		return nil
	case err != nil:
		return err
	}
	if isNonTerminalState(st.State) {
		if _, err := c.SubmitWorkspaceVerb(ctx, id, VerbAbandon); err != nil {
			return err
		}
	}
	if _, err := c.WorkspaceConfig().Delete(ctx, &workspacev1.WorkspaceConfigDeleteRequest{
		Key: workspaceKey(id),
	}); err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}
		return fmt.Errorf("cvp: delete workspace %q: %w", id, err)
	}
	return nil
}

// isNonTerminalState reports whether a workspace can still be abandoned. Once
// SUBMITTED, ABANDONED or ROLLED_BACK it is terminal and abandon is skipped.
func isNonTerminalState(state string) bool {
	switch state {
	case workspacev1.WorkspaceState_WORKSPACE_STATE_PENDING.String(),
		workspacev1.WorkspaceState_WORKSPACE_STATE_CONFLICTS.String(),
		workspacev1.WorkspaceState_WORKSPACE_STATE_UNSPECIFIED.String():
		return true
	default:
		return false
	}
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
