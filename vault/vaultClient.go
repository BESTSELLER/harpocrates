package vault

import (
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/config"
	api "github.com/hashicorp/vault/api"
)

// API is the struct for the vault/api client
type API struct {
	Client *api.Client
}

// NewClient will return a new *API
func NewClient() *API {
	client, err := api.NewClient(&api.Config{
		Address: config.Config.VaultAddress,
	})
	if err != nil {
		fmt.Printf("Unable to create Vault client: %v\n", err)
		os.Exit(1)
	}
	client.SetToken(config.Config.VaultToken)

	return &API{
		Client: client,
	}
}
