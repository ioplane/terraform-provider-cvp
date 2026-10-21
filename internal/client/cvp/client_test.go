package cvp

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	workspacev1 "github.com/ioplane/cvp-sdk/go/gen/arista/workspace.v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
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

func TestNew_TLS(t *testing.T) {
	t.Parallel()
	cert, key := selfSigned(t)

	t.Run("bearer over TLS", func(t *testing.T) {
		t.Parallel()
		client, err := New(Config{Endpoint: "cvp:443", Token: "test", AuthMethod: "bearer"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Cleanup(func() { _ = client.Close() })
	})
	t.Run("valid CA", func(t *testing.T) {
		t.Parallel()
		client, err := New(Config{Endpoint: "cvp:443", Token: "test", AuthMethod: "bearer", CAPEM: cert})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Cleanup(func() { _ = client.Close() })
	})
	t.Run("invalid CA", func(t *testing.T) {
		t.Parallel()
		_, err := New(Config{Endpoint: "cvp:443", Token: "test", AuthMethod: "bearer", CAPEM: []byte("not a pem")})
		if !errors.Is(err, ErrInvalidCACert) {
			t.Fatalf("err = %v, want ErrInvalidCACert", err)
		}
	})
	t.Run("cert mTLS", func(t *testing.T) {
		t.Parallel()
		client, err := New(Config{Endpoint: "cvp:443", AuthMethod: "cert", CertPEM: cert, KeyPEM: key})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Cleanup(func() { _ = client.Close() })
	})
	t.Run("cert bad material", func(t *testing.T) {
		t.Parallel()
		_, err := New(Config{Endpoint: "cvp:443", AuthMethod: "cert", CertPEM: []byte("x"), KeyPEM: []byte("y")})
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
		if c.sdk == nil {
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

// TestProviderUsesSingleProtoOwner catches a return to provider-local wire types.
func TestProviderUsesSingleProtoOwner(t *testing.T) {
	t.Parallel()
	for _, name := range []protoreflect.FullName{"arista.workspace.v1.Workspace", "arista.studio.v1.InputsConfig"} {
		message, err := protoregistry.GlobalTypes.FindMessageByName(name)
		if err != nil {
			t.Fatal(err)
		}
		got := reflect.TypeOf(message.New().Interface()).Elem().PkgPath()
		want := "github.com/ioplane/cvp-sdk/go/gen/" + map[protoreflect.FullName]string{"arista.workspace.v1.Workspace": "arista/workspace.v1", "arista.studio.v1.InputsConfig": "arista/studio.v1"}[name]
		if got != want {
			t.Errorf("%s owner = %s, want %s", name, got, want)
		}
	}
}

type unavailableWorkspace struct {
	workspacev1.UnimplementedWorkspaceConfigServiceServer

	calls     atomic.Int32
	requestID atomic.Value
}

func (s *unavailableWorkspace) Set(_ context.Context, request *workspacev1.WorkspaceConfigSetRequest) (*workspacev1.WorkspaceConfigSetResponse, error) {
	s.calls.Add(1)
	s.requestID.Store(request.GetValue().GetRequestParams().GetRequestId().GetValue())
	return nil, status.Error(codes.PermissionDenied, "request received")
}

func TestSubmitWorkspaceVerbRetainsRequestID(t *testing.T) {
	t.Parallel()
	cert, key := selfSigned(t)
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{pair}})))
	backend := &unavailableWorkspace{}
	workspacev1.RegisterWorkspaceConfigServiceServer(server, backend)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	client, err := New(Config{Endpoint: listener.Addr().String(), AuthMethod: "bearer", Token: "test", InsecureTLS: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	id, err := client.SubmitWorkspaceVerb(ctx, "ws-1", VerbBuild)
	if id == "" || err == nil || !strings.Contains(err.Error(), id) {
		t.Fatalf("request correlation lost: id=%q error=%v", id, err)
	}
	if backend.calls.Load() != 1 || backend.requestID.Load() != id {
		t.Fatalf("submission mismatch: calls=%d sent=%q returned=%q", backend.calls.Load(), backend.requestID.Load(), id)
	}
}
