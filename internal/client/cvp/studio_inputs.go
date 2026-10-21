package cvp

import (
	"context"
	"errors"

	sdk "github.com/ioplane/cvp-sdk/go/cvp"
)

var ErrInputsNotFound = errors.New("cvp: studio inputs not found")

type (
	InputsInput = sdk.InputsInput
	InputsState = sdk.InputsState
)

func (c *Client) SetInputs(ctx context.Context, in InputsInput) error {
	return c.sdk.StudioInputs().Set(ctx, in)
}

func (c *Client) GetInputs(ctx context.Context, studioID, workspaceID string, path []string) (*InputsState, error) {
	value, err := c.sdk.StudioInputs().Get(ctx, studioID, workspaceID, path)
	if errors.Is(err, sdk.ErrNotFound) {
		return nil, ErrInputsNotFound
	}
	return value, err
}

func (c *Client) ListInputPaths(ctx context.Context, studioID, workspaceID string) ([][]string, error) {
	return c.sdk.StudioInputs().ListPaths(ctx, studioID, workspaceID)
}

func (c *Client) DeleteInputs(ctx context.Context, studioID, workspaceID string, path []string) error {
	return c.sdk.StudioInputs().Delete(ctx, studioID, workspaceID, path)
}
