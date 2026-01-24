package vault

import (
	"os"
	"time"

	authAdapter "github.com/BESTSELLER/harpocrates/adapters/secondary/auth"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/rs/zerolog/log"
)

var tokenExpiry time.Time

// Login will exchange the JWT token for a Vault token and only refresh if less than 5 minutes remain
// This function now uses the hexagonal architecture authenticator adapter
func Login() {
	// Create authenticator adapter
	authenticator := authAdapter.NewAdapter(
		config.Config.VaultAddress,
		config.Config.AuthName,
		config.Config.RoleName,
		config.Config.TokenPath,
		config.Config.GcpWorkloadID,
	)

	// Check if token is still valid
	if authenticator.IsTokenValid(config.Config.VaultToken, tokenExpiry) {
		return
	}

	// Perform login
	authResult, err := authenticator.Login()
	if err != nil {
		log.Fatal().Err(err).Msg("Authentication failed")
		os.Exit(1)
	}

	config.Config.VaultToken = authResult.Token
	tokenExpiry = authResult.Expiry
}
