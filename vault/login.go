package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BESTSELLER/go-vault/gcpss"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/token"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

// VaultLoginResult contains the result after logging in.
type VaultLoginResult struct {
	RequestID     string      `json:"request_id"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	LeaseDuration int         `json:"lease_duration"`
	Data          interface{} `json:"data"`
	WrapInfo      interface{} `json:"wrap_info"`
	Warnings      interface{} `json:"warnings"`
	Auth          struct {
		ClientToken   string   `json:"client_token"`
		Accessor      string   `json:"accessor"`
		Policies      []string `json:"policies"`
		TokenPolicies []string `json:"token_policies"`
		Metadata      struct {
			Role                     string `json:"role"`
			ServiceAccountName       string `json:"service_account_name"`
			ServiceAccountNamespace  string `json:"service_account_namespace"`
			ServiceAccountSecretName string `json:"service_account_secret_name"`
			ServiceAccountUID        string `json:"service_account_uid"`
		} `json:"metadata"`
		LeaseDuration int    `json:"lease_duration"`
		Renewable     bool   `json:"renewable"`
		EntityID      string `json:"entity_id"`
		TokenType     string `json:"token_type"`
		Orphan        bool   `json:"orphan"`
	} `json:"auth"`
	Errors []string `json:"errors"`
}

// JWTPayLoad contains the kubernetes token and which role to use
type JWTPayLoad struct {
	Jwt  string `json:"jwt"`
	Role string `json:"role"`
}

// Login will exchange the JWT token to a Vaul token
func Login() {
	if config.Config.VaultToken != "" {
		return
	}
	if config.Config.GcpWorkloadID {
		vaultToken, err := gcpss.FetchVaultToken(config.Config.VaultAddress, config.Config.AuthName)
		if err != nil {
			log.Fatal().Err(err).Msg("GcpWorkload Identity was enabled but auth failed")
			os.Exit(1)
		}
		config.Config.VaultToken = vaultToken
		return
	} else {

		url := config.Config.VaultAddress + "/v1/auth/" + config.Config.AuthName + "/login"

		b := new(bytes.Buffer)
		err := json.NewEncoder(b).Encode(JWTPayLoad{Jwt: token.Read(), Role: config.Config.RoleName})
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to prepare jwt token")
			os.Exit(1)
		}

		req, _ := http.NewRequest("POST", url, b)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to make login call to Vault")
			os.Exit(1)
		}

		returnPayload := VaultLoginResult{}
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
	}
}
