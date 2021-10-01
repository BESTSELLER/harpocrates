package vault

import (
	"os"

	"github.com/BESTSELLER/harpocrates/config"
	api "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
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
		log.Fatal().Err(err).Msg("Unable to create Vault client")
		os.Exit(1)
	}
	client.SetToken(config.Config.VaultToken)

	return &API{
		Client: client,
	}
}
