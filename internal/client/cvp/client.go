package cvp

import (
	"fmt"

	"google.golang.org/grpc"

	workspacev1 "github.com/ioplane/terraform-provider-cvp/internal/pb/arista/workspace.v1"
)

// retryServiceConfig retries only transient failures (design.md D7): UNAVAILABLE
// with exponential backoff, 5 attempts, 1–16 s. Semantic and auth failures
// (AlreadyExists, NotFound, FailedPrecondition, Unauthenticated, PermissionDenied)
// are not retried — they fail fast and are mapped to diagnostics by the caller.
const retryServiceConfig = `{
  "methodConfig": [{
    "name": [{}],
    "retryPolicy": {
      "maxAttempts": 5,
      "initialBackoff": "1s",
      "maxBackoff": "16s",
      "backoffMultiplier": 2.0,
      "retryableStatusCodes": ["UNAVAILABLE"]
    }
  }]
}`

// Config holds everything needed to dial CVP. It is populated from the provider
// configuration block.
type Config struct {
	Endpoint    string
	AuthMethod  string
	Token       string
	CertPEM     []byte
	KeyPEM      []byte
	CAPEM       []byte
	InsecureTLS bool
}

// validate checks the config is internally consistent for the chosen method.
func (c Config) validate() error {
	if c.Endpoint == "" {
		return ErrEndpointRequired
	}
	switch AuthMethod(c.AuthMethod) {
	case AuthBearer, AuthSession:
		if c.Token == "" {
			return ErrTokenRequired
		}
	case AuthCert:
		if len(c.CertPEM) == 0 || len(c.KeyPEM) == 0 {
			return ErrClientCertRequired
		}
	default:
		return ErrUnknownAuthMethod
	}
	return nil
}

// Client is a CVP Resource-API gRPC client. It owns the connection; callers must
// Close it when done (the provider does so on shutdown).
type Client struct {
	conn *grpc.ClientConn
}

// New validates the config and creates a lazy gRPC client (grpc.NewClient does
// not dial until the first RPC). Transport is TLS; bearer/session attach a
// per-RPC Authorization header.
func New(cfg Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	creds, err := buildTransportCreds(cfg)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(retryServiceConfig),
	}
	if tok := cfg.Token; tok != "" && AuthMethod(cfg.AuthMethod) != AuthCert {
		opts = append(opts, grpc.WithPerRPCCredentials(bearerToken{token: tok}))
	}

	conn, err := grpc.NewClient(cfg.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("cvp: create gRPC client for %q: %w", cfg.Endpoint, err)
	}
	return &Client{conn: conn}, nil
}

// Close releases the underlying connection.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Workspace returns the read client (GetOne/GetAll) for workspaces.
func (c *Client) Workspace() workspacev1.WorkspaceServiceClient {
	return workspacev1.NewWorkspaceServiceClient(c.conn)
}

// WorkspaceConfig returns the config (mutating) client — Set creates/updates a
// workspace, Delete removes it.
func (c *Client) WorkspaceConfig() workspacev1.WorkspaceConfigServiceClient {
	return workspacev1.NewWorkspaceConfigServiceClient(c.conn)
}
