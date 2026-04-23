package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/token"
	"github.com/BESTSELLER/harpocrates/vault/gcp"
	"github.com/rs/zerolog/log"
)

// JWTPayLoad contains the kubernetes token and which role to use
type JWTPayLoad struct {
	Jwt  string `json:"jwt"`
	Role string `json:"role"`
}

// Login will exchange the JWT token for a Vault token
func Login() error {
	// If a token exists, validate it. If valid, skip the login.
	if config.Config.VaultToken != "" {
		client := NewClient()
		_, err := client.Client.Auth().Token().LookupSelf()
		if err == nil {
			return nil
		}
		log.Warn().Err(err).Msg("Vault token is invalid or expired, falling back to other authentication methods")
		config.Config.VaultToken = ""
	}

	if config.Config.GcpWorkloadID {
		login, err := gcp.FetchVaultLogin(config.Config.VaultAddress, config.Config.AuthName)
		if err != nil {
			return fmt.Errorf("GcpWorkload Identity was enabled but auth failed: %w", err)
		}
		config.Config.VaultToken = login.Auth.ClientToken
		return nil
	}

	url := config.Config.VaultAddress + "/v1/auth/" + config.Config.AuthName + "/login"

	jwtToken, err := token.Read()
	if err != nil {
		return fmt.Errorf("unable to read token: %w", err)
	}

	payload, err := json.Marshal(JWTPayLoad{Jwt: jwtToken, Role: config.Config.RoleName})
	if err != nil {
		return fmt.Errorf("unable to prepare jwt token: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("unable to create login request to Vault: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make login call to Vault: %w", err)
	}
	defer res.Body.Close() //nolint:errcheck // We don't care about errors from this

	returnPayload := gcp.VaultLoginResult{}
	err = json.NewDecoder(res.Body).Decode(&returnPayload)
	if err != nil {
		return fmt.Errorf("unexpected response from Vault: %w", err)
	}

	if len(returnPayload.Errors) != 0 {
		return fmt.Errorf("API call to Vault failed: %s", returnPayload.Errors)
	}

	config.Config.VaultToken = returnPayload.Auth.ClientToken
	return nil
}
