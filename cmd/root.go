package cmd

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	secretFile string
	// format     string
	// output     string
	// prefix     string
	// secret     string

	rootCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			data := files.ReadFile(secretFile)

			input := util.ReadInput(data)
			allSecrets := util.ExtractSecrets(input)

			if input.Format == "json" {
				files.WriteFile(input.DirPath, fmt.Sprintf("secrets.%s", input.Format), files.FormatAsJSON(allSecrets))
			}

			if input.Format == "env" {
				files.WriteFile(input.DirPath, fmt.Sprintf("secrets.%s", input.Format), files.FormatAsENV(allSecrets))
			}
		},
		Use:   "harpocrates",
		Short: "Will fetch multiple secrets from Hashicorp Vault",
		Long:  "Will fetch multiple secrets from Hashicorp Vault and write them to disk as json, env or secret file",
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
	// rootCmd.PersistentFlags().StringVar(&config.Config.VaultAddress, "vault-address", "", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVar(&config.Config.ClusterName, "cluster-name", "", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVar(&config.Config.TokenPath, "token-path", "", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVar(&config.Config.VaultToken, "vault-token", "", "author name for copyright attribution")

	// rootCmd.PersistentFlags().StringVar(&config.Config.VaultToken, "format", "", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVar(&config.Config.VaultToken, "output", "", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVar(&config.Config.VaultToken, "prefix", "", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVar(&config.Config.VaultToken, "secret", "", "author name for copyright attribution")

}

func initConfig() {
	config.LoadConfig()
	util.GetVaultToken()
}
