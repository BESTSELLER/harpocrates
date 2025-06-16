package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	// "os" // Removed unused import
	"time"

	"github.com/BESTSELLER/go-vault/gcpss"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/token"
	"github.com/rs/zerolog/log"
)

var tokenExpiry time.Time

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

// Login will exchange the JWT token for a Vault token and only refresh if less than 5 minutes remain.
// It returns an error if any step of the login process fails.
func Login() error {
	// tokenIsNotAboutToExpire is true if the token's expiry is more than 5 minutes away.
	tokenIsNotAboutToExpire := time.Now().Add(5 * time.Minute).Before(tokenExpiry)

	// We can reuse the existing token if:
	// 1. Continuous mode is disabled (in this case, we don't proactively refresh based on the 5-minute window).
	// OR
	// 2. Continuous mode is enabled, AND the token is not about to expire within the next 5 minutes.
	canReuseExistingToken := !config.Config.Continuous || tokenIsNotAboutToExpire

	// If a token exists and it meets the conditions for reuse, skip the login.
	if config.Config.VaultToken != "" && canReuseExistingToken {
		return nil
	}

	if config.Config.GcpWorkloadID {
		login, err := gcpss.FetchVaultLogin(config.Config.VaultAddress, config.Config.AuthName)
		if err != nil {
			return fmt.Errorf("GcpWorkload Identity auth failed: %w", err)
		}
		config.Config.VaultToken = login.Auth.ClientToken
		tokenExpiry = time.Now().Add(time.Duration(login.Auth.LeaseDuration) * time.Second)
		return nil
	}

	// Proceed with standard JWT/K8s login
	url := config.Config.VaultAddress + "/v1/auth/" + config.Config.AuthName + "/login"

	jwtToken, err := token.Read()
	if err != nil {
		// token.Read() now returns an error, so we propagate it.
		return fmt.Errorf("failed to read JWT token: %w", err)
	}

	payload := JWTPayLoad{Jwt: jwtToken, Role: config.Config.RoleName}
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(payload); err != nil {
		return fmt.Errorf("unable to prepare JWT payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, b)
	if err != nil {
		return fmt.Errorf("unable to create Vault login request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make login call to Vault: %w", err)
	}
	defer res.Body.Close()

	var returnPayload VaultLoginResult
	if err := json.NewDecoder(res.Body).Decode(&returnPayload); err != nil {
		return fmt.Errorf("unexpected response from Vault, failed to decode JSON: %w", err)
	}

	if len(returnPayload.Errors) != 0 {
		// Consider joining multiple errors if that's a possibility.
		return fmt.Errorf("API call to Vault failed: %v", returnPayload.Errors)
	}

	if returnPayload.Auth.ClientToken == "" {
		return fmt.Errorf("Vault login successful but client token is empty")
	}

	config.Config.VaultToken = returnPayload.Auth.ClientToken
	tokenExpiry = time.Now().Add(time.Duration(returnPayload.Auth.LeaseDuration) * time.Second)
	log.Debug().Msg("Successfully logged into Vault.")
	return nil
}
