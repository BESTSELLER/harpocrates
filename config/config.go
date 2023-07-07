package config

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// GlobalConfig defines the structure of the global configuration parameters
type GlobalConfig struct {
	Append        bool   `required:"false"`
	AuthName      string `required:"false"`
	FileName      string `required:"false"`
	Format        string `required:"false"`
	LogLevel      string `required:"false"`
	Output        string `required:"false"`
	Owner         int    `required:"false"`
	Prefix        string `required:"false"`
	RoleName      string `required:"false"`
	TokenPath     string `required:"false"`
	UpperCase     bool   `required:"false"`
	Validate      bool   `required:"false"`
	VaultAddress  string `required:"false"`
	VaultToken    string `required:"false"`
	GcpWorkloadID bool   `required:"false"`
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
	if Config.AuthName == "" {
		tryEnv("auth_name", &Config.AuthName, required, cmd)
	}
	if Config.RoleName == "" {
		tryEnv("role_name", &Config.RoleName, required, cmd)
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
	if !Config.GcpWorkloadID {
		if envVar, ok := os.LookupEnv("GCP_WORKLOAD_ID"); ok && strings.ToLower(envVar) == "true" {
			(&Config).GcpWorkloadID = true
		}
	}
	if Config.Format == "" {
		tryEnv("format", &Config.Format, notRequired, cmd)
		if Config.Format == "" {
			Config.Format = "env"
		}
	}
	if Config.LogLevel == "" {
		tryEnv("log_level", &Config.LogLevel, notRequired, cmd)
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
