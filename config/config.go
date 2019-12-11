package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// GlobalConfig defines the structure of the global configuration parameters
type GlobalConfig struct {
	VaultAddress string `required:"true" envconfig:"vault_address"`
	ClusterName  string `required:"true" envconfig:"cluster_name"`
	SecretPrefix string `required:"false"`
	VaultToken   string `required:"false"`
}

// Config stores the Global Configuration.
var Config GlobalConfig

//LoadConfig Loads config from env
func LoadConfig() {

	configErr := envconfig.Process("", &Config)
	if configErr != nil {
		log.Fatal(configErr)
	}

}
