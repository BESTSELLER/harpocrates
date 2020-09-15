package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// GlobalConfig defines the structure of the global configuration parameters
type GlobalConfig struct {
	VaultAddress string `required:"false"`
	ClusterName  string `required:"false"`
	TokenPath    string `required:"false"`
	VaultToken   string `required:"false"`
	Prefix       string `required:"false"`
	Format       string `required:"false"`
	Output       string `required:"false"`
	FileName     string `required:"false"`
}

// Config stores the Global Configuration.
var Config GlobalConfig

const (
	notRequired = false
	required    = true
)

// SyncEnvToFlags - We should be able to do this better!
func SyncEnvToFlags(cmd *cobra.Command) {
	if Config.VaultAddress == "" {
		tryEnv("vault_addr", &Config.VaultAddress, required, cmd)
	}
	if Config.ClusterName == "" {
		tryEnv("cluster_name", &Config.ClusterName, required, cmd)
	}
	if Config.TokenPath == "" {
		tryEnv("token_path", &Config.TokenPath, notRequired, cmd)
	}
	if Config.Prefix == "" {
		tryEnv("prefix", &Config.Prefix, notRequired, cmd)
	}
	if Config.VaultToken == "" {
		tryEnv("vault_token", &Config.VaultToken, notRequired, cmd)
	}
	if Config.Format == "" {
		tryEnv("format", &Config.Format, notRequired, cmd)
		if Config.Format == "" {
			Config.Format = "env"
		}

	}
	if Config.FileName == "" {
		tryEnv("HARPOCRATES_FILENAME", &Config.FileName, notRequired, cmd)
		if Config.FileName == "" {
			Config.FileName = "secrets"
		}
	}
}
func tryEnv(env string, some *string, required bool, cmd *cobra.Command) {
	envVar, ok := os.LookupEnv(strings.ToUpper(fmt.Sprintf("%s", env)))
	if ok == true && envVar != "" {
		*some = envVar
	} else {
		if required {
			cmd.MarkPersistentFlagRequired(env)
		}
	}
}
