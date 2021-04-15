package cmd

import (
	"fmt"
	"log"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	secretFile string
	secret     *[]string

	rootCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			var data string
			var input util.SecretJSON

			if secretFile != "" { // --file is being used
				data = files.Read(secretFile)
				input = util.ReadInput(data)
			} else if len(*secret) > 0 { // Parameters is being used
				if config.Config.Output == "" {
					color.Red.Println("Output is required!")
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
				input = util.ReadInput(args[0])
			}

			vault.Login()

			vaultClient := vault.NewClient()

			allSecrets, err := vaultClient.ExtractSecrets(input)
			if err != nil {
				log.Fatal(err)
			}

			if cmd.Flags().Changed("format") && (config.Config.Format != "json" && config.Config.Format != "env" && config.Config.Format != "secret") {
				color.Red.Printf("Please a valid format of either: json, env or secret \n\n")
				cmd.Help()
				return
			}

			for _, v := range allSecrets {
				fileName := config.Config.FileName
				if v.Filename != "" {
					fileName = v.Filename
				}

				if v.Format == "json" {
					files.Write(config.Config.Output, fileName, v.Result.ToJSON(), v.Owner)
				}

				if v.Format == "env" {
					files.Write(config.Config.Output, fileName, v.Result.ToENV(), v.Owner)
				}

				if v.Format == "secret" {
					files.Write(config.Config.Output, fileName, v.Result.ToK8sSecret(), v.Owner)
				}

				color.Green.Printf("Secrets written to file: %s/%s\n", config.Config.Output, fileName)
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

	rootCmd.PersistentFlags().StringVar(&config.Config.Format, "format", "", "output format, either json or env, defaults to env")
	rootCmd.PersistentFlags().StringVar(&config.Config.Output, "output", "", "folder in which secret files will be created e.g. /path/to/folder")
	rootCmd.PersistentFlags().IntVar(&config.Config.Owner, "owner", -1, "UID of the owner that the secret files will be created e.g. 2")
	rootCmd.PersistentFlags().StringVar(&config.Config.Prefix, "prefix", "", "key prefix e.g TEST_ will produce TEST_key=secret")
	rootCmd.PersistentFlags().BoolVar(&config.Config.UpperCase, "uppercase", false, "will convert key to UPPERCASE")
	secret = rootCmd.PersistentFlags().StringSlice("secret", []string{}, "vault path to secret, supports array of secrets e.g. SECRETENGINE/data/test/dev,SECRETENGINE/data/test/prod")

}

func initConfig() {
	config.SyncEnvToFlags(rootCmd)
}
