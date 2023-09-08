package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/BESTSELLER/harpocrates/validate"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/gookit/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
			var data string
			var input util.SecretJSON

			if secretFile != "" { // --file is being used
				data = files.Read(secretFile)
				validFile := validate.SecretsFile(data)
				if !validFile {
					log.Fatal().Msg("Invalid file")
				}
				if config.Config.Validate {
					return
				}
				input = util.ReadInput(data)
			} else if len(*secret) > 0 { // Parameters is being used
				if config.Config.Output == "" {
					log.Error().Msg("Output is required!")
					cmd.Usage()
					return
				}

				y := make([]interface{}, len(*secret))
				// range secrets and assign
				for i, s := range *secret {
					y[i] = s
				}

				input = util.SecretJSON{
					Secrets: y,
				}
			} else { // inline secret is being used
				if len(args) == 0 {
					cmd.Help()
					return
				}

				if validate.SecretsFile(args[0]) {
					input = util.ReadInput(args[0])
				}
				if config.Config.Validate {
					return
				}
			}

			continuous := false
			envVar, ok := os.LookupEnv("CONTINUOUS")
			if ok && strings.ToLower(envVar) == "true" {
				continuous = true

				interval, _ := os.LookupEnv("INTERVAL")

				durationParsed, err := time.ParseDuration(interval)
				if err != nil {
					log.Fatal().Err(err).Msgf("%s", err)
				}

				if durationParsed < (1 * time.Minute) {
					log.Fatal().Msg("Interval must be at least 1 minute")
				}

				duration = &durationParsed
				log.Debug().Msgf("Continuous mode enabled, will run every %s", durationParsed)

				// If we are in continuous mode, we want to overwrite to the file
				config.Config.Append = false
				http.Handle("/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if success {
						w.WriteHeader(http.StatusOK)
					} else {
						w.WriteHeader(http.StatusTooEarly)
					}
				}))
				go func() {
					http.ListenAndServe(":8000", nil)
				}()
			}

			for {
				vault.Login()

				vaultClient := vault.NewClient()

				allSecrets, err := vaultClient.ExtractSecrets(input, config.Config.Append)
				if err != nil {
					success = false
					log.Fatal().Err(err).Msgf("%s", err)
				}

				if cmd.Flags().Changed("format") && (config.Config.Format != "json" && config.Config.Format != "env" && config.Config.Format != "secret") {
					log.Error().Msg("Please a valid format of either: json, env or secret")
					cmd.Help()
					return
				}

				for _, v := range allSecrets {
					fileName := config.Config.FileName
					if v.Filename != "" {
						fileName = v.Filename
					}

					if v.Format == "json" {
						files.Write(config.Config.Output, fileName, v.Result.ToJSON(), v.Owner, config.Config.Append)
					} else if v.Format == "env" {
						files.Write(config.Config.Output, fileName, v.Result.ToENV(), v.Owner, config.Config.Append)
					} else if v.Format == "secret" {
						files.Write(config.Config.Output, fileName, v.Result.ToK8sSecret(), v.Owner, config.Config.Append)
					} else if v.Format == "yaml" || v.Format == "yml" {
						files.Write(config.Config.Output, fileName, v.Result.ToYAML(), v.Owner, config.Config.Append)
					}
					log.Debug().Msgf("Secrets written to file: %s/%s", config.Config.Output, fileName)
				}
				success = true

				if !continuous {
					break
				}

				log.Debug().Msgf("Sleeping for %s", duration)
				time.Sleep(*duration)
			}

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
