package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/gookit/color"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	duration *time.Duration
	// Used for flags.
	secretFile string
	secret     *[]string
	success    bool

	rootCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			doIt(cmd, args, false)
		},
		Use:   "harpocrates",
		Short: fmt.Sprintf("%sThis application will fetch secrets from Hashicorp Vault", color.Blue.Sprintf("\"Harpocrates was the god of silence, secrets and confidentiality\"\n")),
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Setup flags
	rootCmd.PersistentFlags().StringVarP(&secretFile, "file", "f", "", "that contains the configuration to apply")
	rootCmd.PersistentFlags().StringVar(&config.Config.VaultAddress, "vault-address", "", "url to vault e.g. https://vault.example.com")
	rootCmd.PersistentFlags().StringVar(&config.Config.AuthName, "auth-name", "", "k8s auth method name, use when running as a k8s pod")
	rootCmd.PersistentFlags().StringVar(&config.Config.RoleName, "role-name", "", "k8s auth role name, use when running as a k8s pod")
	rootCmd.PersistentFlags().StringVar(&config.Config.TokenPath, "token-path", "", "/path/to/token/file")
	rootCmd.PersistentFlags().StringVar(&config.Config.VaultToken, "vault-token", "", "vault token in clear text")
	rootCmd.PersistentFlags().BoolVar(&config.Config.GcpWorkloadID, "gcpWorkloadID", false, "Enable GcpWorkloadID auth method instead of using vault token")

	rootCmd.PersistentFlags().StringVar(&config.Config.Format, "format", "", "output format, either json or env, defaults to env")
	rootCmd.PersistentFlags().StringVar(&config.Config.Output, "output", "", "folder in which secret files will be created e.g. /path/to/folder")
	rootCmd.PersistentFlags().IntVar(&config.Config.Owner, "owner", -1, "UID of the owner that the secret files will be created e.g. 2")
	rootCmd.PersistentFlags().StringVar(&config.Config.Prefix, "prefix", "", "key prefix e.g TEST_ will produce TEST_key=secret")
	rootCmd.PersistentFlags().BoolVar(&config.Config.UpperCase, "uppercase", false, "will convert key to UPPERCASE")
	rootCmd.PersistentFlags().StringVar(&config.Config.LogLevel, "log-level", "", "LogLevel, default is warn")
	rootCmd.PersistentFlags().BoolVar(&config.Config.Validate, "validate", false, "Validate, will only validate the secrets file")
	rootCmd.PersistentFlags().BoolVar(&config.Config.Append, "append", true, "Append, appends secrets to a file, defaults to true")
	secret = rootCmd.PersistentFlags().StringSlice("secret", []string{}, "vault path to secret, supports array of secrets e.g. SECRETENGINE/data/test/dev,SECRETENGINE/data/test/prod")

}

func initConfig() {
	config.SyncEnvToFlags(rootCmd)
	SetupLogLevel()
}

// SetupLogLevel sets the global loglevel
func SetupLogLevel() {
	// default is info
	switch strings.ToLower(config.Config.LogLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
