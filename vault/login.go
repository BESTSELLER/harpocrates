package vault

import (
	"os"
	"sync"
	"time"

	authAdapter "github.com/BESTSELLER/harpocrates/adapters/secondary/auth"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/domain/ports"
	"github.com/rs/zerolog/log"
)

var (
	tokenExpiry       time.Time
	authenticator     ports.Authenticator
	authenticatorOnce sync.Once
	loginMu           sync.Mutex
)

// getAuthenticator returns a cached authenticator, creating it once if needed.
// Note: The authenticator is initialized with the config values at first access.
// If config.Config.Continuous changes after initialization, the cached authenticator
// will continue using the initial value. This is by design for simplicity, as
// continuous mode is typically set once at application startup.
func getAuthenticator() ports.Authenticator {
	authenticatorOnce.Do(func() {
		authenticator = authAdapter.NewAdapter(
			config.Config.VaultAddress,
			config.Config.AuthName,
			config.Config.RoleName,
			config.Config.TokenPath,
			config.Config.GcpWorkloadID,
			config.Config.Continuous,
		)
	})
	return authenticator
}

// Login will exchange the JWT token for a Vault token and only refresh if less than 5 minutes remain.
// This function now uses the hexagonal architecture authenticator adapter.
// It is safe for concurrent access.
func Login() {
	loginMu.Lock()
	defer loginMu.Unlock()
	
	auth := getAuthenticator()

	// Check if token is still valid
	if auth.IsTokenValid(config.Config.VaultToken, tokenExpiry) {
		return
	}

	// Perform login
	authResult, err := auth.Login()
	if err != nil {
		log.Fatal().Err(err).Msg("Authentication failed")
		os.Exit(1)
	}

	config.Config.VaultToken = authResult.Token
	tokenExpiry = authResult.Expiry
}
