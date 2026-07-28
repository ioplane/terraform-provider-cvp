package cvp

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	studiov1 "github.com/ioplane/terraform-provider-cvp/internal/pb/arista/studio.v1"
	fmp "github.com/ioplane/terraform-provider-cvp/internal/pb/fmp"
)

// ErrInputsNotFound is returned by GetInputs when no inputs value exists at the
// given studio path in the workspace (gRPC NotFound). Callers map it to state
// removal.
var ErrInputsNotFound = errors.New("cvp: studio inputs not found")

// InputsInput is the declarative desired state of a studio inputs entry — the
// config the cvp_studio_inputs resource converges via InputsConfigService.Set.
// InputsJSON is the JSON-encoded value at Path; an empty Path targets the studio
// root. CVP validates the value against the studio schema on Set.
type InputsInput struct {
	StudioID    string
	WorkspaceID string
	Path        []string
	InputsJSON  string
}

// InputsState is the read-back view of a studio inputs entry.
type InputsState struct {
	StudioID    string
	WorkspaceID string
	Path        []string
	InputsJSON  string
}

func inputsKey(studioID, workspaceID string, path []string) *studiov1.InputsKey {
	return &studiov1.InputsKey{
		StudioId:    wrapperspb.String(studioID),
		WorkspaceId: wrapperspb.String(workspaceID),
		Path:        &fmp.RepeatedString{Values: path},
	}
}

// SetInputs creates or updates the inputs value at a studio path in a workspace.
func (c *Client) SetInputs(ctx context.Context, in InputsInput) error {
	_, err := c.InputsConfig().Set(ctx, &studiov1.InputsConfigSetRequest{
		Value: &studiov1.InputsConfig{
			Key:    inputsKey(in.StudioID, in.WorkspaceID, in.Path),
			Inputs: wrapperspb.String(in.InputsJSON),
		},
	})
	if err != nil {
		return fmt.Errorf("cvp: set studio inputs (studio %q, path %v): %w", in.StudioID, in.Path, err)
	}
	return nil
}

// GetInputs reads the config value at a studio path. The config view round-trips
// exactly what SetInputs wrote (unlike the resolved InputsService state view). A
// NotFound is translated to ErrInputsNotFound.
func (c *Client) GetInputs(ctx context.Context, studioID, workspaceID string, path []string) (*InputsState, error) {
	resp, err := c.InputsConfig().GetOne(ctx, &studiov1.InputsConfigRequest{
		Key: inputsKey(studioID, workspaceID, path),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrInputsNotFound
		}
		return nil, fmt.Errorf("cvp: get studio inputs (studio %q, path %v): %w", studioID, path, err)
	}
	v := resp.GetValue()
	return &InputsState{
		StudioID:    v.GetKey().GetStudioId().GetValue(),
		WorkspaceID: v.GetKey().GetWorkspaceId().GetValue(),
		Path:        v.GetKey().GetPath().GetValues(),
		InputsJSON:  v.GetInputs().GetValue(),
	}, nil
}

// DeleteInputs removes the inputs config entry at a studio path in the
// workspace. A missing entry is treated as already gone.
func (c *Client) DeleteInputs(ctx context.Context, studioID, workspaceID string, path []string) error {
	_, err := c.InputsConfig().Delete(ctx, &studiov1.InputsConfigDeleteRequest{
		Key: inputsKey(studioID, workspaceID, path),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}
		return fmt.Errorf("cvp: delete studio inputs (studio %q, path %v): %w", studioID, path, err)
	}
	return nil
}
