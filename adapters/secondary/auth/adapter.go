package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BESTSELLER/harpocrates/domain/ports"
	"github.com/BESTSELLER/harpocrates/token"
	"github.com/BESTSELLER/harpocrates/vault/gcp"
)

// JWTPayLoad contains the kubernetes token and which role to use
type JWTPayLoad struct {
	Jwt  string `json:"jwt"`
	Role string `json:"role"`
}

// Adapter implements the Authenticator port for Vault authentication
type Adapter struct {
	vaultAddress  string
	authName      string
	roleName      string
	tokenPath     string
	gcpWorkloadID bool
	continuous    bool
}

// NewAdapter creates a new auth adapter
func NewAdapter(vaultAddress, authName, roleName, tokenPath string, gcpWorkloadID bool, continuous bool) ports.Authenticator {
	return &Adapter{
		vaultAddress:  vaultAddress,
		authName:      authName,
		roleName:      roleName,
		tokenPath:     tokenPath,
		gcpWorkloadID: gcpWorkloadID,
		continuous:    continuous,
	}
}

// Login authenticates and returns a token
func (a *Adapter) Login() (*ports.AuthResult, error) {
	if a.gcpWorkloadID {
		return a.loginWithGCP()
	}

	return a.loginWithKubernetes()
}

// IsTokenValid checks if the current token is still valid
func (a *Adapter) IsTokenValid(token string, expiry time.Time) bool {
	if token == "" {
		return false
	}

	// Token is valid if it's not about to expire within the next 5 minutes
	tokenIsNotAboutToExpire := time.Now().Add(5 * time.Minute).Before(expiry)

	// We can reuse the existing token if:
	// 1. Continuous mode is disabled (in this case, we don't proactively refresh based on the 5-minute window).
	// OR
	// 2. Continuous mode is enabled, AND the token is not about to expire within the next 5 minutes.
	return !a.continuous || tokenIsNotAboutToExpire
}

func (a *Adapter) loginWithGCP() (*ports.AuthResult, error) {
	login, err := gcp.FetchVaultLogin(a.vaultAddress, a.authName)
	if err != nil {
		return nil, fmt.Errorf("GCP workload identity auth failed: %w", err)
	}

	return &ports.AuthResult{
		Token:         login.Auth.ClientToken,
		Expiry:        time.Now().Add(time.Duration(login.Auth.LeaseDuration) * time.Second),
		LeaseDuration: login.Auth.LeaseDuration,
	}, nil
}

func (a *Adapter) loginWithKubernetes() (*ports.AuthResult, error) {
	url := a.vaultAddress + "/v1/auth/" + a.authName + "/login"

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(JWTPayLoad{Jwt: token.Read(), Role: a.roleName})
	if err != nil {
		return nil, fmt.Errorf("unable to prepare jwt token: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, b)
	if err != nil {
		return nil, fmt.Errorf("unable to create login request to Vault: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make login call to Vault: %w", err)
	}
	defer res.Body.Close()

	returnPayload := gcp.VaultLoginResult{}
	err = json.NewDecoder(res.Body).Decode(&returnPayload)
	if err != nil {
		return nil, fmt.Errorf("unexpected response from Vault: %w", err)
	}

	if len(returnPayload.Errors) != 0 {
		return nil, fmt.Errorf("API call to Vault failed: %s", returnPayload.Errors)
	}

	return &ports.AuthResult{
		Token:         returnPayload.Auth.ClientToken,
		Expiry:        time.Now().Add(time.Duration(returnPayload.Auth.LeaseDuration) * time.Second),
		LeaseDuration: returnPayload.Auth.LeaseDuration,
	}, nil
}
