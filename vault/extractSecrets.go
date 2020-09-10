package vault

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/secrets"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/mitchellh/mapstructure"
)

// ExtractSecrets will loop through al those damn interfaces
func (vaultClient *API) ExtractSecrets(input util.SecretJSON) (secrets.Result, error) {
	var result = make(secrets.Result)
	var currentPrefix = config.Config.Prefix

	for _, a := range input.Secrets {

		// If the key is just a secret path, then it will read that from Vault, otherwise:
		if fmt.Sprintf("%T", a) != "string" {
			b, ok := a.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected map[string]interface{}, got: '%s'", a)
			}

			aa := map[string]util.Secret{}
			mapstructure.Decode(b, &aa)

			for c, d := range aa {
				setPrefix(d.Prefix, &currentPrefix)

				for _, f := range d.Keys {

					// If the key is just a secret path, then it will read that from Vault, otherwise:
					if fmt.Sprintf("%T", f) != "string" {
						bb := map[string]util.SecretKeys{}
						mapstructure.Decode(f, &bb)

						for h, i := range bb {
							setPrefix(i.Prefix, &currentPrefix)

							if i.SaveAsFile != nil {
								secretValue, err := vaultClient.ReadSecretKey(fmt.Sprintf("%s", c), h)
								if err != nil {
									return nil, err
								}
								if *i.SaveAsFile {
									fmt.Println("Creating file...", h)
									files.Write(input.Output, fmt.Sprintf("%s%s", currentPrefix, h), secretValue)
								} else {
									result.Add(h, secretValue, currentPrefix)
								}
							} else {
								secretValue, err := vaultClient.ReadSecretKey(fmt.Sprintf("%s", c), h)
								if err != nil {
									return nil, err
								}
								result.Add(h, secretValue, currentPrefix)
							}
							setPrefix(d.Prefix, &currentPrefix)
						}
					} else {
						secretValue, err := vaultClient.ReadSecretKey(fmt.Sprintf("%s", c), fmt.Sprintf("%s", f))
						if err != nil {
							return nil, err
						}
						result.Add(fmt.Sprintf("%s", f), secretValue, currentPrefix)
					}
				}
				setPrefix(config.Config.Prefix, &currentPrefix)
			}
		} else {
			secretValue, err := vaultClient.ReadSecret(fmt.Sprintf("%s", a))
			if err != nil {
				return nil, err
			}
			for aa, bb := range secretValue {
				result.Add(aa, fmt.Sprintf("%s", bb), currentPrefix)
			}
		}
	}
	return result, nil
}

func setPrefix(potentialPrefix string, currentPrefix *string) {
	if potentialPrefix != "" {
		*currentPrefix = potentialPrefix
	} else {
		*currentPrefix = config.Config.Prefix
	}
}