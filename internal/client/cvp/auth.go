package cvp

// AuthMethod selects the shared SDK identity method.
type AuthMethod string

const (
	AuthCert    AuthMethod = "cert"
	AuthSession AuthMethod = "session"
	AuthBearer  AuthMethod = "bearer"
)
