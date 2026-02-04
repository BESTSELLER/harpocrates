package vault

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/secrets"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/go-viper/mapstructure/v2"
)

type Outputs struct {
	Format   string         `json:"format,omitempty"    yaml:"format,omitempty"`
	Filename string         `json:"filename,omitempty"  yaml:"filename,omitempty"`
	Result   secrets.Result `json:"result,omitempty"    yaml:"result,omitempty"`
	Owner    *int           `json:"owner,omitempty"     yaml:"owner,omitempty"`
}

// ExtractSecrets will loop through al those damn interfaces
func (vaultClient *API) ExtractSecrets(input util.SecretJSON, appendToFile bool) ([]Outputs, error) {
	var finalResult []Outputs
	var result = make(secrets.Result)
	var currentPrefix = config.Config.Prefix
	var currentUpperCase = config.Config.UpperCase
	var currentFormat = config.Config.Format

	for _, a := range input.Secrets {

		// If the key is just a secret path, then it will read that from Vault, otherwise:
		if secretPath, ok := util.GetSecretString(a); ok {
			// Handle simple string secret path
			secretValue, err := vaultClient.ReadSecret(secretPath)
			if err != nil {
				return nil, err
			}
			for aa, bb := range secretValue {
				result.Add(aa, bb, currentPrefix, currentUpperCase)
			}
		} else if secretMap, ok := util.GetSecretMap(a); ok {
			// Handle Secret struct (as map)
			aa := map[string]util.Secret{}
			mapstructure.Decode(secretMap, &aa)

			for c, d := range aa {
				setPrefix(d.Prefix, &currentPrefix)
				setUpper(d.UpperCase, &currentUpperCase)
				setFormat(d.Format, &currentFormat)

				if len(d.Keys) == 0 {
					secretValue, err := vaultClient.ReadSecret(c)
					if err != nil {
						return nil, err
					}
					var thisResult = make(secrets.Result)
					for k, v := range secretValue {
						thisResult.Add(k, v, currentPrefix, currentUpperCase)
					}

					finalResult = append(finalResult, Outputs{Format: currentFormat, Filename: d.FileName, Result: thisResult, Owner: d.Owner})
					continue
				}

				for _, f := range d.Keys {
					// If the key is just a string, then it will read that key from Vault, otherwise:
					if keyName, ok := util.GetKeyString(f); ok {
						// Handle simple string key name
						secretValue, err := vaultClient.ReadSecretKey(c, keyName)
						if err != nil {
							return nil, err
						}
						result.Add(keyName, secretValue, currentPrefix, currentUpperCase)
					} else if keyMap, ok := util.GetKeyMap(f); ok {
						// Handle SecretKeys struct (as map)
						bb := map[string]util.SecretKeys{}
						mapstructure.Decode(keyMap, &bb)

						for h, i := range bb {
							setPrefix(i.Prefix, &currentPrefix)
							setUpper(d.UpperCase, &currentUpperCase)

							if i.SaveAsFile != nil {
								secretValue, err := vaultClient.ReadSecretKey(c, h)
								if err != nil {
									return nil, err
								}
								if *i.SaveAsFile {
									files.Write(input.Output, secrets.ToUpperOrNotToUpper(fmt.Sprintf("%s%s", currentPrefix, h), &currentUpperCase), secretValue, nil, appendToFile)
								} else {
									result.Add(h, secretValue, currentPrefix, currentUpperCase)
								}
							} else {
								secretValue, err := vaultClient.ReadSecretKey(c, h)
								if err != nil {
									return nil, err
								}
								result.Add(h, secretValue, currentPrefix, currentUpperCase)
							}
							setPrefix(d.Prefix, &currentPrefix)
							setUpper(d.UpperCase, &currentUpperCase)
						}
					} else {
						return nil, fmt.Errorf("invalid key item type: expected string or SecretKeys map, got: %T", f)
					}
				}
				setPrefix(config.Config.Prefix, &currentPrefix)
				setUpper(d.UpperCase, &currentUpperCase)
				setFormat(d.Format, &currentFormat)
			}
		} else {
			return nil, fmt.Errorf("invalid secret item type: expected string or Secret map, got: %T", a)
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

func setFormat(potentiaFormat string, currentFormat *string) {
	if potentiaFormat != "" {
		*currentFormat = potentiaFormat
	} else {
		*currentFormat = config.Config.Format
	}
}
