package vault_test

import (
	"testing"

	vaultadapter "github.com/BESTSELLER/harpocrates/adapters/secondary/vault"
	"github.com/BESTSELLER/harpocrates/domain/ports"
	api "github.com/hashicorp/vault/api"
)

// TestVaultAdapterImplementsPort verifies that the Vault adapter implements the SecretFetcher port
func TestVaultAdapterImplementsPort(t *testing.T) {
	client, _ := api.NewClient(&api.Config{Address: "http://localhost:8200"})
	var _ ports.SecretFetcher = vaultadapter.NewAdapter(client)
}
