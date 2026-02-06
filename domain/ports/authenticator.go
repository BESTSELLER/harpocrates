package ports

import "time"

// AuthResult contains the result of authentication
type AuthResult struct {
	Token         string
	Expiry        time.Time
	LeaseDuration int
}

// Authenticator defines the port for authenticating with secret stores
type Authenticator interface {
	// Login authenticates and returns a token
	Login() (*AuthResult, error)
	
	// IsTokenValid checks if the current token is still valid
	IsTokenValid(token string, expiry time.Time) bool
}
