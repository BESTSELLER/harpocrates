package vault

import (
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/hashicorp/vault/api"
)

func createClient() *api.Client {
	client, err := api.NewClient(&api.Config{
		Address: config.Config.VaultAddress,
	})
	client.Token()
	if err != nil {
		fmt.Printf("Unable to create Vault client: %v\n", err)
		os.Exit(1)
	}
	client.SetToken(config.Config.VaultToken)
	return client
}
