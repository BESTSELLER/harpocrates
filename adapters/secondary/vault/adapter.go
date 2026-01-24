package vault

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/BESTSELLER/harpocrates/domain/ports"
	api "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
)

// Adapter implements the SecretFetcher port for HashiCorp Vault
type Adapter struct {
	client *api.Client
}

// NewAdapter creates a new Vault adapter
func NewAdapter(client *api.Client) ports.SecretFetcher {
	return &Adapter{
		client: client,
	}
}

// ReadSecret reads all key-value pairs from a secret path
func (a *Adapter) ReadSecret(path string) (map[string]interface{}, error) {
	log.Debug().Msgf("Reading secret from path: %s", path)
	secretValues, err := a.client.Logical().Read(path)
	if secretValues == nil {
		return nil, fmt.Errorf("the secret '%s' was not found: %v", path, err)
	}

	secretData := secretValues.Data["data"]

	if secretData == nil {
		secretData = secretValues.Data
	}

	// Will append data to path and retry if "data" is empty and warnings is present
	// if path contains data and warnings an error is returned
	if fmt.Sprintf("%s", secretValues.Data) == fmt.Sprintf("%s", make(map[string]interface{})) {
		if len(secretValues.Warnings) > 0 {
			splitPath := strings.Split(path, "/")
			if splitPath[1] == "data" {
				return nil, fmt.Errorf("%s", strings.Join(secretValues.Warnings, ","))
			}

			appendData := []string{splitPath[0], "data"}
			pathWithData := append(appendData, splitPath[1:]...)
			return a.ReadSecret(strings.Join(pathWithData, "/"))
		}
		return nil, fmt.Errorf("no data recieved")
	}

	b, err := json.Marshal(secretData)
	if err != nil {
		return nil, err
	}

	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response from Vault")
	}

	myMap := f.(map[string]interface{})

	return myMap, nil
}

// ReadSecretKey reads a specific key from a secret path
// Supports nested key access using dot notation (e.g., "globalSecrets.theSecretINeed")
// and array indexing (e.g., "list[0]", "users[0].name")
func (a *Adapter) ReadSecretKey(path string, key string) (interface{}, error) {
	log.Debug().Msgf("Reading secret key '%s' from path: %s", key, path)
	secret, err := a.ReadSecret(path)
	if secret == nil {
		return "", fmt.Errorf("the key '%s' was not found in the path '%s': %v", key, path, err)
	}
	if err != nil {
		return "", err
	}

	// 1. Literal match
	if val, ok := secret[key]; ok {
		return val, nil
	}

	// 2. Traversal with nested keys and array access
	normalizedKey := strings.ReplaceAll(key, "[", ".")
	normalizedKey = strings.ReplaceAll(normalizedKey, "]", "")
	keys := strings.Split(normalizedKey, ".")

	var current interface{} = secret
	for _, k := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[k]; exists {
				current = val
				continue
			}
		}
		if a, ok := current.([]interface{}); ok {
			if i, err := strconv.Atoi(k); err == nil {
				if i >= 0 && i < len(a) {
					current = a[i]
					continue
				}
			}
		}
		return "", fmt.Errorf("the key '%s' was not found in the path '%s': %v", key, path, nil)
	}
	return current, nil
}
