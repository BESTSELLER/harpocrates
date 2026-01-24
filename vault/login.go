package vault

import (
	"os"
	"time"

	authAdapter "github.com/BESTSELLER/harpocrates/adapters/secondary/auth"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/domain/ports"
	"github.com/rs/zerolog/log"
)

var (
	tokenExpiry    time.Time
	authenticator  ports.Authenticator
)

// getAuthenticator returns a cached authenticator or creates a new one if config has changed
func getAuthenticator() ports.Authenticator {
	// Create new authenticator only if it doesn't exist yet
	// In a real DI setup, this would be injected once at startup
	if authenticator == nil {
		authenticator = authAdapter.NewAdapter(
			config.Config.VaultAddress,
			config.Config.AuthName,
			config.Config.RoleName,
			config.Config.TokenPath,
			config.Config.GcpWorkloadID,
			config.Config.Continuous,
		)
	}
	return authenticator
}

// Login will exchange the JWT token for a Vault token and only refresh if less than 5 minutes remain
// This function now uses the hexagonal architecture authenticator adapter
func Login() {
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
