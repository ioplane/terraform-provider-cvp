package cvp

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	sdk "github.com/ioplane/cvp-sdk/go/cvp"
	"github.com/ioplane/cvp-sdk/go/cvp/auth"
)

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

// Client adapts the shared SDK to the provider's existing internal interfaces.
type Client struct{ sdk *sdk.Client }

// New validates provider configuration and creates the shared authenticated client.
func New(cfg Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13}
	// Explicit practitioner opt-in; the SDK preserves TLS encryption.
	tlsConfig.InsecureSkipVerify = cfg.InsecureTLS
	if len(cfg.CAPEM) > 0 {
		tlsConfig.RootCAs = x509.NewCertPool()
		if !tlsConfig.RootCAs.AppendCertsFromPEM(cfg.CAPEM) {
			return nil, ErrInvalidCACert
		}
	}
	var identity auth.Method
	switch AuthMethod(cfg.AuthMethod) {
	case AuthBearer:
		identity = auth.Bearer(cfg.Token)
	case AuthSession:
		identity = auth.Session(cfg.Token)
	case AuthCert:
		identity = auth.Cert(cfg.CertPEM, cfg.KeyPEM, nil)
	}
	client, err := sdk.NewClient(cfg.Endpoint, sdk.WithAuth(identity), sdk.WithTLSConfig(tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("cvp: create SDK client: %w", err)
	}
	return &Client{sdk: client}, nil
}

// Close releases the shared SDK connection.
func (c *Client) Close() error { return c.sdk.Close() }
