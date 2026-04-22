package vault

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/secrets"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/go-viper/mapstructure/v2"
)

// Outputs represents the output format for secrets
type Outputs struct {
	Format   string         `json:"format,omitempty"    yaml:"format,omitempty"`
	Filename string         `json:"filename,omitempty"  yaml:"filename,omitempty"`
	Result   secrets.Result `json:"result,omitempty"    yaml:"result,omitempty"`
	Owner    *int           `json:"owner,omitempty"     yaml:"owner,omitempty"`
}

// ExtractSecrets will loop through all the provided secret interfaces
func (vaultClient *API) ExtractSecrets(input util.SecretJSON, appendToFile bool) ([]Outputs, error) {
	var finalResult []Outputs
	var result = make(secrets.Result)
	var currentPrefix = config.Config.Prefix
	var currentUpperCase = config.Config.UpperCase
	var currentFormat = config.Config.Format

	for _, secretEntry := range input.Secrets {

		// If the key is just a secret path, then it will read that from Vault, otherwise:
		if _, isString := secretEntry.(string); !isString {
			secretMapRaw, ok := secretEntry.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected map[string]interface{}, got: '%s'", secretEntry)
			}

			secretConfigMap := map[string]util.Secret{}
			err := mapstructure.Decode(secretMapRaw, &secretConfigMap)
			if err != nil {
				return finalResult, err
			}

			for secretPath, secretConfig := range secretConfigMap {
				setPrefix(secretConfig.Prefix, &currentPrefix)
				setUpper(secretConfig.UpperCase, &currentUpperCase)
				setFormat(secretConfig.Format, &currentFormat)

				if len(secretConfig.Keys) == 0 {
					secretValue, err := vaultClient.ReadSecret(secretPath)
					if err != nil {
						return nil, err
					}
					var thisResult = make(secrets.Result)
					for key, value := range secretValue {
						thisResult.Add(key, value, currentPrefix, currentUpperCase)
					}

					finalResult = append(finalResult, Outputs{Format: currentFormat, Filename: secretConfig.FileName, Result: thisResult, Owner: secretConfig.Owner})
					continue
				}

				for _, keyEntry := range secretConfig.Keys {
					// If the key is just a secret path, then it will read that from Vault, otherwise:
					if _, isString := keyEntry.(string); !isString {
						secretKeysConfigMap := map[string]util.SecretKeys{}
						err := mapstructure.Decode(keyEntry, &secretKeysConfigMap)
						if err != nil {
							return finalResult, err
						}

						for vaultKey, keyConfig := range secretKeysConfigMap {
							setPrefix(keyConfig.Prefix, &currentPrefix)
							setUpper(keyConfig.UpperCase, &currentUpperCase)

							keyName := vaultKey
							if keyConfig.Alias != "" {
								keyName = keyConfig.Alias
							}

							if keyConfig.SaveAsFile != nil {
								secretValue, err := vaultClient.ReadSecretKey(secretPath, vaultKey)
								if err != nil {
									return nil, err
								}
								if *keyConfig.SaveAsFile {
									files.Write(input.Output, secrets.ToUpperOrNotToUpper(fmt.Sprintf("%s%s", currentPrefix, keyName), &currentUpperCase), secretValue, nil, appendToFile)
								} else {
									result.Add(keyName, secretValue, currentPrefix, currentUpperCase)
								}
							} else {
								secretValue, err := vaultClient.ReadSecretKey(secretPath, vaultKey)
								if err != nil {
									return nil, err
								}
								result.Add(keyName, secretValue, currentPrefix, currentUpperCase)
							}
							setPrefix(secretConfig.Prefix, &currentPrefix)
							setUpper(secretConfig.UpperCase, &currentUpperCase)
						}
					} else {
						secretValue, err := vaultClient.ReadSecretKey(secretPath, fmt.Sprintf("%s", keyEntry))
						if err != nil {
							return nil, err
						}
						result.Add(fmt.Sprintf("%s", keyEntry), secretValue, currentPrefix, currentUpperCase)
					}
				}
				setPrefix(config.Config.Prefix, &currentPrefix)
				setUpper(secretConfig.UpperCase, &currentUpperCase)
				setFormat(secretConfig.Format, &currentFormat)
			}
		} else {
			secretValue, err := vaultClient.ReadSecret(fmt.Sprintf("%s", secretEntry))
			if err != nil {
				return nil, err
			}
			for key, value := range secretValue {
				result.Add(key, value, currentPrefix, currentUpperCase)
			}
		}
	}

	finalResult = append(finalResult, Outputs{Format: config.Config.Format, Filename: "", Result: result})
	return finalResult, nil
}

func setPrefix(potentialPrefix string, currentPrefix *string) {
	if potentialPrefix != "" {
		*currentPrefix = potentialPrefix
	} else {
		*currentPrefix = config.Config.Prefix
	}
}
func setUpper(potentialUpper *bool, currentUpper *bool) {
	if potentialUpper != nil {
		*currentUpper = *potentialUpper
	} else {
		*currentUpper = config.Config.UpperCase
	}
}

func setFormat(potentialFormat string, currentFormat *string) {
	if potentialFormat != "" {
		*currentFormat = potentialFormat
	} else {
		*currentFormat = config.Config.Format
	}
}
