//go:build acceptance

// Package acceptance holds the live acceptance suite (terraform-plugin-testing),
// gated by the `acceptance` build tag and TF_ACC=1. It runs real
// plan/apply/refresh/import/destroy cycles against the netlab2 CVP lab and needs
// CVP_ENDPOINT / CVP_AUTH_METHOD / CVP_TOKEN (see docs/testing.md). Never point
// it at a production cluster.
package acceptance

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/ioplane/terraform-provider-cvp/internal/client/cvp"
	"github.com/ioplane/terraform-provider-cvp/internal/provider"
)

// protoV6ProviderFactories serves the provider in-process for the test harness.
var protoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cvp": providerserver.NewProtocol6WithError(provider.New("acc")()),
}

// preCheck fails fast when the lab credentials are absent.
func preCheck(t *testing.T) {
	t.Helper()
	for _, k := range []string{"CVP_ENDPOINT", "CVP_AUTH_METHOD", "CVP_TOKEN"} {
		if os.Getenv(k) == "" {
			t.Fatalf("%s must be set for acceptance tests", k)
		}
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// providerConfig renders a provider block from the lab environment.
func providerConfig() string {
	return fmt.Sprintf(`
provider "cvp" {
  endpoint     = %q
  auth_method  = %q
  token        = %q
  insecure_tls = %s
}
`, os.Getenv("CVP_ENDPOINT"), os.Getenv("CVP_AUTH_METHOD"), os.Getenv("CVP_TOKEN"), envOr("CVP_INSECURE_TLS", "false"))
}

// accClient builds a direct CVP client from the environment for CheckDestroy and
// the action-verb lifecycle test.
func accClient(t *testing.T) *cvp.Client {
	t.Helper()
	client, err := cvp.New(cvp.Config{
		Endpoint:    os.Getenv("CVP_ENDPOINT"),
		AuthMethod:  os.Getenv("CVP_AUTH_METHOD"),
		Token:       os.Getenv("CVP_TOKEN"),
		InsecureTLS: envOr("CVP_INSECURE_TLS", "false") == "true",
	})
	if err != nil {
		t.Fatalf("build acceptance CVP client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// workspaceGone reports whether a workspace no longer exists, surfacing any
// transport error to the caller.
func workspaceGone(ctx context.Context, client *cvp.Client, id string) (bool, error) {
	_, err := client.GetWorkspace(ctx, id)
	switch {
	case err == nil:
		return false, nil
	case errors.Is(err, cvp.ErrWorkspaceNotFound):
		return true, nil
	default:
		return false, err
	}
}
