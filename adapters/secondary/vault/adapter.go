package vault

import (
	"fmt"

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
	secret, err := a.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret from path '%s': %w", path, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("no secret found at path '%s'", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		return secret.Data, nil
	}

	secretData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected secret format at path '%s'", path)
	}

	return secretData, nil
}

// ReadSecretKey reads a specific key from a secret path
func (a *Adapter) ReadSecretKey(path string, key string) (string, error) {
	log.Debug().Msgf("Reading secret key '%s' from path: %s", key, path)
	secretData, err := a.ReadSecret(path)
	if err != nil {
		return "", err
	}

	value, ok := secretData[key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found in secret at path '%s'", key, path)
	}

	return fmt.Sprintf("%v", value), nil
}
