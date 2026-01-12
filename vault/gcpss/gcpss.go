package gcpss

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
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

type data struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Data          struct {
		Data     interface{} `json:"data"`
		Metadata struct {
			CreatedTime  time.Time `json:"created_time"`
			DeletionTime string    `json:"deletion_time"`
			Destroyed    bool      `json:"destroyed"`
			Version      int       `json:"version"`
		} `json:"metadata"`
	} `json:"data"`
	WrapInfo interface{} `json:"wrap_info"`
	Warnings interface{} `json:"warnings"`
	Auth     interface{} `json:"auth"`
}

func fetchJWT(vaultRole string) (jwt string, err error) {
	client := metadata.NewClient(http.DefaultClient)
	return client.GetWithContext(context.Background(), "instance/service-accounts/default/identity?audience=http://vault/"+vaultRole+"&format=full")
}

func fetchVaultToken(vaultAddr string, jwt string, vaultRole string) (vaultToken string, err error) {
	login, err := fetchVaultLogin(vaultAddr, jwt, vaultRole)
	if err != nil {
		return "", err
	}

	return login.Auth.ClientToken, nil
}

func fetchVaultLogin(vaultAddr string, jwt string, vaultRole string) (VaultLoginResult, error) {
	var login VaultLoginResult
	client := http.DefaultClient

	j := `{"role":"` + vaultRole + `", "jwt":"` + jwt + `"}`

	req, err := http.NewRequest(http.MethodPost, vaultAddr+"/v1/auth/gcp/login", bytes.NewBufferString(j))
	if err != nil {
		return login, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return login, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&login)
	if err != nil {
		return login, err
	}

	if len(login.Errors) > 0 {
		return login, fmt.Errorf("%s", login.Errors[0])
	}
	if login.Auth.ClientToken == "" {
		return login, fmt.Errorf("unable to retrieve vault token")
	}
	if resp.StatusCode < 200 || resp.StatusCode > 202 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return login, err
		}
		return login, fmt.Errorf("request failed, expected status: 2xx got: %d, error message %s", resp.StatusCode, string(body))
	}

	return login, nil
}

func readSecret(vaultAddr string, vaultToken string, vaultSecret string) (secret string, err error) {
	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, vaultAddr+"/v1/"+vaultSecret, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Vault-Token", vaultToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var s data

	if decoderErr := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return "", decoderErr
	}

	data, err := json.Marshal(s.Data.Data)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 202 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("request failed, expected status: 2xx got: %d, error message %s", resp.StatusCode, string(body))
	}

	return string(data), nil
}

// FetchVaultToken gets Workload Identity Token from GCP Metadata API and uses it to fetch Vault Token.
func FetchVaultToken(vaultAddr string, vaultRole string) (vaultToken string, err error) {
	jwt, err := fetchJWT(vaultRole)
	if err != nil {
		return "", err
	}

	token, err := fetchVaultToken(vaultAddr, jwt, vaultRole)
	if err != nil {
		return "", err
	}

	return token, nil
}

// FetchVaultLogin gets Workload Identity Token from GCP Metadata API and uses it to fetch Vault Login object.
func FetchVaultLogin(vaultAddr string, vaultRole string) (VaultLoginResult, error) {
	jwt, err := fetchJWT(vaultRole)
	if err != nil {
		return VaultLoginResult{}, err
	}

	login, err := fetchVaultLogin(vaultAddr, jwt, vaultRole)
	if err != nil {
		return VaultLoginResult{}, err
	}

	return login, nil
}

// FetchVaultSecret returns secret from Hashicorp Vault.
func FetchVaultSecret(vaultAddr string, vaultSecret string, vaultRole string) (secret string, err error) {
	token, err := FetchVaultToken(vaultAddr, vaultRole)
	if err != nil {
		return "", err
	}

	data, err := readSecret(vaultAddr, token, vaultSecret)
	if err != nil {
		return "", err
	}
	return data, nil
}
