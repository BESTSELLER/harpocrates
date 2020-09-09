package vault

import (
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/config"
	vaultapi "github.com/hashicorp/vault/api"
)

// Client declares a switchable api.Client
// var Client *api.Client

// func createClient() *api.Client {

// 	if Client == nil {

// 		client, err := api.NewClient(&api.Config{
// 			Address: config.Config.VaultAddress,
// 		})
// 		client.Token()
// 		if err != nil {
// 			fmt.Printf("Unable to create Vault Client: %v\n", err)
// 			os.Exit(1)
// 		}
// 		client.SetToken(config.Config.VaultToken)
// 		Client = client
// 	}
// 	fmt.Println(Client)
// 	return Client
// }

type API struct {
	Client *vaultapi.Client
}

func (client *API) Create() *vaultapi.Client {
	if client == nil {
		tempClient, err := vaultapi.NewClient(&vaultapi.Config{
			Address: config.Config.VaultAddress,
		})
		tempClient.Token()
		if err != nil {
			fmt.Printf("Unable to create Vault client: %v\n", err)
			os.Exit(1)
		}
		tempClient.SetToken(config.Config.VaultToken)
		client = &API{
			Client: tempClient,
		}
	}
	return client.Client
}
