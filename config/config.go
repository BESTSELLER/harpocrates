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
	tryEnv("vault_addr", &Config.VaultAddress, required, cmd)
	tryEnv("auth_name", &Config.AuthName, required, cmd)
	tryEnv("role_name", &Config.RoleName, required, cmd)
	tryEnv("token_path", &Config.TokenPath, notRequired, cmd)
	tryEnv("prefix", &Config.Prefix, notRequired, cmd)
	tryEnv("vault_token", &Config.VaultToken, notRequired, cmd)
	if !Config.GcpWorkloadID {
		if envVar, ok := os.LookupEnv("GCP_WORKLOAD_ID"); ok && strings.ToLower(envVar) == "true" {
			(&Config).GcpWorkloadID = true
		}
	}
	tryEnv("format", &Config.Format, notRequired, cmd)
	if Config.Format == "" {
		Config.Format = "env"
	}
	tryEnv("log_level", &Config.LogLevel, notRequired, cmd)
	tryEnv("HARPOCRATES_FILENAME", &Config.FileName, notRequired, cmd)
	if Config.FileName == "" {
		Config.FileName = "secrets"
	}
}
func tryEnv(env string, some *string, required bool, cmd *cobra.Command) {
	if *some != "" {
		return
	}
	envVar, ok := os.LookupEnv(strings.ToUpper(fmt.Sprintf("%s", env)))
	if ok == true && envVar != "" {
		*some = envVar
	} else {
		if required {
			_ = cmd.MarkPersistentFlagRequired(env)
		}
	}
}
