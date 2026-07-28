package cvp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"
)

// selfSigned returns a matching cert/key PEM pair for TLS tests.
func selfSigned(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "cvp-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM
}

func TestConfig_validate(t *testing.T) {
	t.Parallel()
	cert, key := selfSigned(t)
	cases := []struct {
		name string
		cfg  Config
		want error
	}{
		{"bearer ok", Config{Endpoint: "cvp:443", AuthMethod: "bearer", Token: "t"}, nil},
		{"session ok", Config{Endpoint: "cvp:443", AuthMethod: "session", Token: "t"}, nil},
		{"cert ok", Config{Endpoint: "cvp:443", AuthMethod: "cert", CertPEM: cert, KeyPEM: key}, nil},
		{"no endpoint", Config{AuthMethod: "bearer", Token: "t"}, ErrEndpointRequired},
		{"unknown method", Config{Endpoint: "cvp:443", AuthMethod: "kerberos"}, ErrUnknownAuthMethod},
		{"bearer no token", Config{Endpoint: "cvp:443", AuthMethod: "bearer"}, ErrTokenRequired},
		{"cert no material", Config{Endpoint: "cvp:443", AuthMethod: "cert"}, ErrClientCertRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.cfg.validate(); !errors.Is(err, tc.want) {
				t.Fatalf("validate() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	t.Parallel()
	b := bearerToken{token: "abc123"}
	md, err := b.GetRequestMetadata(t.Context())
	if err != nil {
		t.Fatalf("GetRequestMetadata: %v", err)
	}
	if got := md["authorization"]; got != "Bearer abc123" {
		t.Fatalf("authorization = %q, want %q", got, "Bearer abc123")
	}
	if !b.RequireTransportSecurity() {
		t.Fatal("RequireTransportSecurity() = false, want true")
	}
}

func TestBuildTransportCreds(t *testing.T) {
	t.Parallel()
	cert, key := selfSigned(t)

	t.Run("bearer over TLS", func(t *testing.T) {
		t.Parallel()
		if _, err := buildTransportCreds(Config{AuthMethod: "bearer"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("valid CA", func(t *testing.T) {
		t.Parallel()
		if _, err := buildTransportCreds(Config{AuthMethod: "bearer", CAPEM: cert}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("invalid CA", func(t *testing.T) {
		t.Parallel()
		_, err := buildTransportCreds(Config{AuthMethod: "bearer", CAPEM: []byte("not a pem")})
		if !errors.Is(err, ErrInvalidCACert) {
			t.Fatalf("err = %v, want ErrInvalidCACert", err)
		}
	})
	t.Run("cert mTLS", func(t *testing.T) {
		t.Parallel()
		if _, err := buildTransportCreds(Config{AuthMethod: "cert", CertPEM: cert, KeyPEM: key}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("cert bad material", func(t *testing.T) {
		t.Parallel()
		_, err := buildTransportCreds(Config{AuthMethod: "cert", CertPEM: []byte("x"), KeyPEM: []byte("y")})
		if err == nil {
			t.Fatal("expected error for invalid cert material")
		}
	})
}

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("bearer", func(t *testing.T) {
		t.Parallel()
		c, err := New(Config{Endpoint: "cvp.example.com:443", AuthMethod: "bearer", Token: "t"})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		t.Cleanup(func() { _ = c.Close() })
		if c.WorkspaceConfig() == nil || c.Workspace() == nil {
			t.Fatal("service clients must be non-nil")
		}
	})
	t.Run("missing token", func(t *testing.T) {
		t.Parallel()
		if _, err := New(Config{Endpoint: "cvp:443", AuthMethod: "bearer"}); !errors.Is(err, ErrTokenRequired) {
			t.Fatalf("err = %v, want ErrTokenRequired", err)
		}
	})
}
