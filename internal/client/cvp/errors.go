// Package cvp is the gRPC client for Arista CloudVision Portal Resource APIs.
//
// Authentication follows the CVP authz model (evidence:
// arista-cvp-re/docs/integration-drafts/um-docs/07-architecture/cvp/03-authz-model.md):
// the apiserver runs with `-authn=cert,session,certhdr` and
// `enablebearertokenlogin=true`. Three methods are supported here, matching
// design.md D4:
//
//   - bearer  — an API/service-account token (minted via
//     arista.serviceaccount.v1.TokenConfigService) sent as the gRPC metadata
//     `authorization: Bearer <token>`. Recommended for CI/CD.
//   - session — the `access_token` JWT from
//     `/cvpservice/login/authenticate.do`; because the apiserver enables
//     bearer-token login, the session JWT is accepted as a bearer token too, so
//     it travels the same metadata path.
//   - cert    — client mTLS.
//
// Transport is gRPC/TLS to the CVP public listener (Ambassador/nginx :443).
package cvp

import "errors"

var (
	// ErrEndpointRequired is returned when Config.Endpoint is empty.
	ErrEndpointRequired = errors.New("cvp: endpoint is required")
	// ErrUnknownAuthMethod is returned for an auth_method outside cert/session/bearer.
	ErrUnknownAuthMethod = errors.New("cvp: unknown auth_method (want cert, session or bearer)")
	// ErrTokenRequired is returned when session/bearer auth is selected without a token.
	ErrTokenRequired = errors.New("cvp: token is required for session/bearer auth")
	// ErrClientCertRequired is returned when cert auth is selected without cert_pem/key_pem.
	ErrClientCertRequired = errors.New("cvp: cert_pem and key_pem are required for cert auth")
	// ErrInvalidCACert is returned when the provided CA PEM cannot be parsed.
	ErrInvalidCACert = errors.New("cvp: ca_pem is not a valid PEM certificate bundle")
	// ErrUnknownWorkspaceVerb is returned by SubmitWorkspaceVerb for a verb
	// outside the known WorkspaceVerb set.
	ErrUnknownWorkspaceVerb = errors.New("cvp: unknown workspace verb")
)
