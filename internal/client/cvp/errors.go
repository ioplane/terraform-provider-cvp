// Package cvp adapts the shared SDK to the existing Terraform provider contract.
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
)
