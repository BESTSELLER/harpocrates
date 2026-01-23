package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/token"
	"github.com/BESTSELLER/harpocrates/vault/gcp"
	"github.com/rs/zerolog/log"
)

var tokenExpiry time.Time

// JWTPayLoad contains the kubernetes token and which role to use
type JWTPayLoad struct {
	Jwt  string `json:"jwt"`
	Role string `json:"role"`
}

// Login will exchange the JWT token for a Vault token and only refresh if less than 5 minutes remain
func Login() {
	// tokenIsNotAboutToExpire is true if the token's expiry is more than 5 minutes away.
	tokenIsNotAboutToExpire := time.Now().Add(5 * time.Minute).Before(tokenExpiry)

	// We can reuse the existing token if:
	// 1. Continuous mode is disabled (in this case, we don't proactively refresh based on the 5-minute window).
	// OR
	// 2. Continuous mode is enabled, AND the token is not about to expire within the next 5 minutes.
	canReuseExistingToken := !config.Config.Continuous || tokenIsNotAboutToExpire

	// If a token exists and it meets the conditions for reuse, skip the login.
	if config.Config.VaultToken != "" && canReuseExistingToken {
		return
	}

	if config.Config.GcpWorkloadID {
		login, err := gcp.FetchVaultLogin(config.Config.VaultAddress, config.Config.AuthName)
		if err != nil {
			log.Fatal().Err(err).Msg("GcpWorkload Identity was enabled but auth failed")
			os.Exit(1)
		}
		config.Config.VaultToken = login.Auth.ClientToken
		tokenExpiry = time.Now().Add(time.Duration(login.Auth.LeaseDuration) * time.Second)
		return
	}

	url := config.Config.VaultAddress + "/v1/auth/" + config.Config.AuthName + "/login"

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(JWTPayLoad{Jwt: token.Read(), Role: config.Config.RoleName})
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to prepare jwt token")
		os.Exit(1)
	}

	req, _ := http.NewRequest(http.MethodPost, url, b)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to make login call to Vault")
		os.Exit(1)
	}
	defer res.Body.Close()

	returnPayload := gcp.VaultLoginResult{}
	err = json.NewDecoder(res.Body).Decode(&returnPayload)
	if err != nil {
		log.Fatal().Err(err).Msg("Unexpected response from Vault")
		os.Exit(1)
	}

	if len(returnPayload.Errors) != 0 {
		log.Fatal().Err(fmt.Errorf("%s", returnPayload.Errors)).Msg("API call to Vault failed")
		os.Exit(1)
	}

	config.Config.VaultToken = returnPayload.Auth.ClientToken
	tokenExpiry = time.Now().Add(time.Duration(returnPayload.Auth.LeaseDuration) * time.Second)
}
