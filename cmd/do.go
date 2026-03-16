package cmd

import (
	"os"
	"path"
	"strings"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/BESTSELLER/harpocrates/validate"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func loadLocalVaultToken() {
	if config.Config.VaultToken != "" {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Debug().Err(err).Msg("unable to get home directory")
		return
	}

	vaultTokenFile := path.Join(homeDir, ".vault-token")
	token, err := os.ReadFile(vaultTokenFile)
	if err == nil {
		config.Config.VaultToken = strings.TrimSpace(string(token))
		log.Debug().Msg("using vault token from ~/.vault-token")
	}

}

func doIt(cmd *cobra.Command, args []string) []string {
	loadLocalVaultToken()

	secretEnvs := []string{}

	var data string
	var input util.SecretJSON

	if secretFile != "" {
		data = files.Read(secretFile)
		validFile := validate.SecretsFile(data)
		if !validFile {
			log.Fatal().Msg("Invalid file")
		}
		if config.Config.Validate {
			return secretEnvs
		}
		input = util.ReadInput(data)
	} else if len(*secret) > 0 {
		if config.Config.Output == "" {
			log.Error().Msg("Output is required!")
			cmd.Usage() //nolint:errcheck // We don't care about errors from this
			return secretEnvs
		}

		y := make([]interface{}, len(*secret))

		for i, s := range *secret {
			y[i] = s
		}

		input = util.SecretJSON{
			Secrets: y,
		}
	} else {
		if len(args) == 0 {
			cmd.Help() //nolint:errcheck // We don't care about errors from this
			return secretEnvs
		}

		if validate.SecretsFile(args[0]) {
			input = util.ReadInput(args[0])
		}
		if config.Config.Validate {
			return secretEnvs
		}
	}

	vault.Login()

	vaultClient := vault.NewClient()

	allSecrets, err := vaultClient.ExtractSecrets(input, config.Config.Append)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to extract secrets from Vault")
	}

	if cmd.Flags().Changed("format") && (config.Config.Format != "json" && config.Config.Format != "env" && config.Config.Format != "secret" && config.Config.Format != "yaml") {
		log.Error().Msg("Please use a valid format of either: json, env, secret or yaml")
		cmd.Help() //nolint:errcheck // We don't care about errors from this
		return secretEnvs
	}

	for _, v := range allSecrets {
		fileName := config.Config.FileName
		if v.Filename != "" {
			fileName = v.Filename
		}

		switch v.Format {
		case "json":
			files.Write(config.Config.Output, fileName, v.Result.ToJSON(), v.Owner, config.Config.Append)
		case "env":
			files.Write(config.Config.Output, fileName, v.Result.ToENV(), v.Owner, config.Config.Append)
			secretEnvs = append(secretEnvs, v.Result.ToKVarray("")...)
		case "secret":
			files.Write(config.Config.Output, fileName, v.Result.ToK8sSecret(), v.Owner, config.Config.Append)
		case "yaml":
			files.Write(config.Config.Output, fileName, v.Result.ToYAML(), v.Owner, config.Config.Append)
		}
		log.Debug().Msgf("Secrets written to file: %s/%s", config.Config.Output, fileName)
	}

	return secretEnvs
}
