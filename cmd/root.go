package cmd

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/spf13/cobra"
	"gopkg.in/gookit/color.v1"
)

var (
	// Used for flags.
	secretFile string
	secret     string

	rootCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			var data string
			var input util.SecretJSON

			if secretFile != "" { // --file is being used
				data = files.Read(secretFile)
				input = util.ReadInput(data)
			} else if secret != "" { // Parameters is being used
				if config.Config.Output == "" {
					color.Red.Println("Output is required!")
					cmd.Usage()
					return
				}

				y := make([]interface{}, 1)
				y[0] = secret

				input = util.SecretJSON{
					Secrets: y,
				}
			} else { // inline secret is being used
				if len(args) == 0 {
					cmd.Help()
					return
				}
				input = util.ReadInput(args[0])
			}

			vault.Login()
			allSecrets := util.ExtractSecrets(input)
			fileName := fmt.Sprintf("secrets.%s", config.Config.Format)

			if config.Config.Format == "json" {
				files.Write(config.Config.Output, fileName, allSecrets.ToJSON())
			}

			if config.Config.Format == "env" {
				files.Write(config.Config.Output, fileName, allSecrets.ToENV())
			}
			color.Green.Printf("Secrets written to file: %s/%s\n", config.Config.Output, fileName)
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
	rootCmd.PersistentFlags().StringVar(&config.Config.VaultAddress, "vault-address", "", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVar(&config.Config.ClusterName, "cluster-name", "", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVar(&config.Config.TokenPath, "token-path", "", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVar(&config.Config.VaultToken, "vault-token", "", "author name for copyright attribution")

	rootCmd.PersistentFlags().StringVar(&config.Config.Format, "format", "", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVar(&config.Config.Output, "output", "", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVar(&config.Config.Prefix, "prefix", "", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVar(&secret, "secret", "", "author name for copyright attribution")

}

func initConfig() {
	config.SyncEnvToFlags(rootCmd)
}
