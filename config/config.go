package config

import (
	"os"
	"strings"

	"github.com/rs/zerolog" // Added import here
	"github.com/spf13/cobra"
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
	Continuous    bool   `required:"false"`
}

// Config stores the Global Configuration.
var Config GlobalConfig
var secretFile string
var secret *[]string

const (
	notRequired = false
	required    = true
)

// InitFlags initializes the command line flags
func InitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&secretFile, "file", "f", "", "that contains the configuration to apply")
	cmd.PersistentFlags().StringVar(&Config.VaultAddress, "vault-address", "", "url to vault e.g. https://vault.example.com")
	cmd.PersistentFlags().StringVar(&Config.AuthName, "auth-name", "", "k8s auth method name, use when running as a k8s pod")
	cmd.PersistentFlags().StringVar(&Config.RoleName, "role-name", "", "k8s auth role name, use when running as a k8s pod")
	cmd.PersistentFlags().StringVar(&Config.TokenPath, "token-path", "", "/path/to/token/file")
	cmd.PersistentFlags().StringVar(&Config.VaultToken, "vault-token", "", "vault token in clear text")
	cmd.PersistentFlags().BoolVar(&Config.GcpWorkloadID, "gcpWorkloadID", false, "Enable GcpWorkloadID auth method instead of using vault token")

	cmd.PersistentFlags().StringVar(&Config.Format, "format", "", "output format, either json or env, defaults to env")
	cmd.PersistentFlags().StringVar(&Config.Output, "output", "", "folder in which secret files will be created e.g. /path/to/folder")
	cmd.PersistentFlags().IntVar(&Config.Owner, "owner", -1, "UID of the owner that the secret files will be created e.g. 2")
	cmd.PersistentFlags().StringVar(&Config.Prefix, "prefix", "", "key prefix e.g TEST_ will produce TEST_key=secret")
	cmd.PersistentFlags().BoolVar(&Config.UpperCase, "uppercase", false, "will convert key to UPPERCASE")
	cmd.PersistentFlags().StringVar(&Config.LogLevel, "log-level", "", "LogLevel, default is warn")
	cmd.PersistentFlags().BoolVar(&Config.Validate, "validate", false, "Validate, will only validate the secrets file")
	cmd.PersistentFlags().BoolVar(&Config.Append, "append", true, "Append, appends secrets to a file, defaults to true")
	secret = cmd.PersistentFlags().StringSlice("secret", []string{}, "vault path to secret, supports array of secrets e.g. SECRETENGINE/data/test/dev,SECRETENGINE/data/test/prod")
}

// SyncEnvToFlags syncs environment variables to flags and configuration.
func SyncEnvToFlags(cmd *cobra.Command) {
	tryEnv("vault_addr", &Config.VaultAddress, required, cmd, "vault-address")
	tryEnv("auth_name", &Config.AuthName, required, cmd, "auth-name")
	tryEnv("role_name", &Config.RoleName, required, cmd, "role-name")
	tryEnv("token_path", &Config.TokenPath, notRequired, cmd, "token-path")
	tryEnv("prefix", &Config.Prefix, notRequired, cmd, "prefix")
	tryEnv("vault_token", &Config.VaultToken, notRequired, cmd, "vault-token")
	if !Config.GcpWorkloadID {
		if envVar, ok := os.LookupEnv("GCP_WORKLOAD_ID"); ok && strings.ToLower(envVar) == "true" {
			Config.GcpWorkloadID = true
		}
	}
	tryEnv("format", &Config.Format, notRequired, cmd, "format")
	if Config.Format == "" {
		Config.Format = "env" // Default format
	}
	tryEnv("log_level", &Config.LogLevel, notRequired, cmd, "log-level")
	// HARPOCRATES_FILENAME is a direct environment variable, not a flag.
	// It sets Config.FileName directly.
	if envVar, ok := os.LookupEnv("HARPOCRATES_FILENAME"); ok {
		Config.FileName = envVar
	}
	if Config.FileName == "" {
		Config.FileName = "secrets" // Default filename
	}

	// Handle CONTINUOUS environment variable
	if envVar, ok := os.LookupEnv("CONTINUOUS"); ok && strings.ToLower(envVar) == "true" {
		Config.Continuous = true
	}
	// Note: INTERVAL is handled in cmd/root.go due to time.ParseDuration logic.
}

// tryEnv attempts to set a configuration value from an environment variable if not already set by a flag.
// It now also accepts the corresponding flag name to mark it as required if the environment variable is missing.
func tryEnv(envName string, configField *string, isRequired bool, cmd *cobra.Command, flagName string) {
	// If the config field is already set (e.g., by a flag), do nothing.
	if configField != nil && *configField != "" {
		return
	}

	envVarValue, envVarSet := os.LookupEnv(strings.ToUpper(envName))
	if envVarSet && envVarValue != "" {
		*configField = envVarValue
	} else if isRequired {
		// If the environment variable is not set but is required, mark the corresponding flag as required.
		// This relies on the flag having been defined elsewhere (e.g., in InitFlags).
		if flag := cmd.PersistentFlags().Lookup(flagName); flag != nil {
			_ = cmd.MarkPersistentFlagRequired(flagName)
		}
	}
}

// SetupLogLevel sets the global log level based on Config.LogLevel.
func SetupLogLevel() {
	switch strings.ToLower(Config.LogLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // Default to InfoLevel
	}
}

// GetSecretFile returns the secret file path.
func GetSecretFile() string {
	return secretFile
}

// GetSecretSlice returns the slice of secrets.
func GetSecretSlice() []string {
	if secret == nil {
		return []string{}
	}
	return *secret
}
