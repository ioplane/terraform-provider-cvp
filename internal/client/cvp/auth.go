package cvp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc/credentials"
)

// AuthMethod is one of the three CVP authentication methods (design.md D4).
type AuthMethod string

const (
	AuthCert    AuthMethod = "cert"
	AuthSession AuthMethod = "session"
	AuthBearer  AuthMethod = "bearer"
)

// bearerToken is a per-RPC credential that sends `authorization: Bearer <token>`
// on every call. Used for both the bearer and session methods (the CVP apiserver
// accepts a session JWT as a bearer token — see the package comment).
type bearerToken struct {
	token string
}

func (b bearerToken) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + b.token}, nil
}

// RequireTransportSecurity is true: the token must never leave over a plaintext
// transport. The provider always dials over TLS (even the opt-in insecure_tls
// path uses TLS with verification disabled), so this holds.
func (b bearerToken) RequireTransportSecurity() bool { return true }

// buildTransportCreds derives the TLS transport credentials from the config:
// TLS 1.3 minimum, an optional pinned CA bundle, and a client certificate for
// the cert (mTLS) method.
func buildTransportCreds(cfg Config) (credentials.TransportCredentials, error) {
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS13}

	if cfg.InsecureTLS {
		// Opt-in, documented lab-only flag; never a default. See SECURITY.md.
		// (gosec's G402 matches the composite-literal form, not this assignment;
		// CodeQL's equivalent alert is dismissed as intentional.)
		tlsCfg.InsecureSkipVerify = true
	}

	if len(cfg.CAPEM) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(cfg.CAPEM) {
			return nil, ErrInvalidCACert
		}
		tlsCfg.RootCAs = pool
	}

	if AuthMethod(cfg.AuthMethod) == AuthCert {
		cert, err := tls.X509KeyPair(cfg.CertPEM, cfg.KeyPEM)
		if err != nil {
			return nil, fmt.Errorf("cvp: load client certificate: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return credentials.NewTLS(tlsCfg), nil
}
