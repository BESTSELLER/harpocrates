package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// GlobalConfig defines the structure of the global configuration parameters
type GlobalConfig struct {
	VaultAddress string `required:"false" envconfig:"vault_address"`
	ClusterName  string `required:"false" envconfig:"cluster_name"`
	TokenPath    string `required:"false" envconfig:"token_path"`
	SecretPrefix string `required:"false"`
	VaultToken   string `required:"false"`
}

// Config stores the Global Configuration.
var Config GlobalConfig

//LoadConfig Loads config from env
func LoadConfig() {

	configErr := envconfig.Process("harpocrates", &Config)
	if configErr != nil {
		log.Fatal(configErr)
	}

}
