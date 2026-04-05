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
func Login() {
	// If a token exists, skip the login.
	if config.Config.VaultToken != "" {
		return
	}

	if config.Config.GcpWorkloadID {
		login, err := gcp.FetchVaultLogin(config.Config.VaultAddress, config.Config.AuthName)
		if err != nil {
			log.Fatal().Err(err).Msg("GcpWorkload Identity was enabled but auth failed")
		}
		config.Config.VaultToken = login.Auth.ClientToken
		return
	}

	url := config.Config.VaultAddress + "/v1/auth/" + config.Config.AuthName + "/login"

	payload, err := json.Marshal(JWTPayLoad{Jwt: token.Read(), Role: config.Config.RoleName})
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to prepare jwt token")
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create login request to Vault")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to make login call to Vault")
	}
	defer res.Body.Close() //nolint:errcheck // We don't care about errors from this

	returnPayload := gcp.VaultLoginResult{}
	err = json.NewDecoder(res.Body).Decode(&returnPayload)
	if err != nil {
		log.Fatal().Err(err).Msg("Unexpected response from Vault")
	}

	if len(returnPayload.Errors) != 0 {
		log.Fatal().Err(fmt.Errorf("%s", returnPayload.Errors)).Msg("API call to Vault failed")
	}

	config.Config.VaultToken = returnPayload.Auth.ClientToken
}
