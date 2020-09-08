package util

import (
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/secrets"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/mitchellh/mapstructure"
)

// ExtractSecrets will loop through al those damn interfaces
func ExtractSecrets(input SecretJSON) secrets.Result {
	var result = make(secrets.Result)
	var currentPrefix = config.Config.Prefix

	for _, a := range input.Secrets {

		// If the key is just a secret path, then it will read that from Vault, otherwise:
		if fmt.Sprintf("%T", a) != "string" {
			b, ok := a.(map[string]interface{})
			if !ok {
				fmt.Printf("Expected map[string]interface{}, got: '%s'\n", a)
				os.Exit(1)
			}

			aa := map[string]Secret{}
			mapstructure.Decode(b, &aa)

			for c, d := range aa {
				setPrefix(d.Prefix, &currentPrefix)

				for _, f := range d.Keys {

					// If the key is just a secret path, then it will read that from Vault, otherwise:
					if fmt.Sprintf("%T", f) != "string" {
						bb := map[string]SecretKeys{}
						mapstructure.Decode(f, &bb)

						for h, i := range bb {
							setPrefix(i.Prefix, &currentPrefix)

							if i.SaveAsFile != nil {
								var secretValue = vault.ReadSecretKey(fmt.Sprintf("%s", c), h)
								if *i.SaveAsFile {
									fmt.Println("Creating file...", h)
									files.Write(input.Output, fmt.Sprintf("%s%s", currentPrefix, h), secretValue)
								} else {
									result.Add(h, secretValue, currentPrefix)
								}
							} else {
								var secretValue = vault.ReadSecretKey(fmt.Sprintf("%s", c), h)
								result.Add(h, secretValue, currentPrefix)
							}
							setPrefix(d.Prefix, &currentPrefix)
						}
					} else {
						var secretValue = vault.ReadSecretKey(fmt.Sprintf("%s", c), fmt.Sprintf("%s", f))
						result.Add(fmt.Sprintf("%s", f), secretValue, currentPrefix)
					}
				}
			}
		} else {
			var secretValue = vault.ReadSecret(fmt.Sprintf("%s", a))
			for aa, bb := range secretValue {
				result.Add(aa, fmt.Sprintf("%s", bb), currentPrefix)
			}
		}
	}
	return result
}

func setPrefix(potentialPrefix string, currentPrefix *string) {
	if potentialPrefix != "" {
		*currentPrefix = potentialPrefix
	} else {
		*currentPrefix = config.Config.Prefix
	}
}
