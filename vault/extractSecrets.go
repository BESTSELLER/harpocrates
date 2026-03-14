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

// ExtractSecrets will loop through al those damn interfaces
func (vaultClient *API) ExtractSecrets(input util.SecretJSON, appendToFile bool) ([]Outputs, error) {
	var finalResult []Outputs
	var result = make(secrets.Result)
	var currentPrefix = config.Config.Prefix
	var currentUpperCase = config.Config.UpperCase
	var currentFormat = config.Config.Format

	for _, a := range input.Secrets {

		// If the key is just a secret path, then it will read that from Vault, otherwise:
		if fmt.Sprintf("%T", a) != "string" {
			b, ok := a.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected map[string]interface{}, got: '%s'", a)
			}

			aa := map[string]util.Secret{}
			err := mapstructure.Decode(b, &aa)
			if err != nil {
				return finalResult, err
			}

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
					// If the key is just a secret path, then it will read that from Vault, otherwise:
					if fmt.Sprintf("%T", f) != "string" {
						bb := map[string]util.SecretKeys{}
						err := mapstructure.Decode(f, &bb)
						if err != nil {
							return finalResult, err
						}

						for h, i := range bb {
							setPrefix(i.Prefix, &currentPrefix)
							setUpper(d.UpperCase, &currentUpperCase)

							keyName := h
							if i.OverrideName != "" {
								keyName = i.OverrideName
							}

							if i.SaveAsFile != nil {
								secretValue, err := vaultClient.ReadSecretKey(c, h)
								if err != nil {
									return nil, err
								}
								if *i.SaveAsFile {
									files.Write(input.Output, secrets.ToUpperOrNotToUpper(fmt.Sprintf("%s%s", currentPrefix, keyName), &currentUpperCase), secretValue, nil, appendToFile)
								} else {
									result.Add(keyName, secretValue, currentPrefix, currentUpperCase)
								}
							} else {
								secretValue, err := vaultClient.ReadSecretKey(c, h)
								if err != nil {
									return nil, err
								}
								result.Add(keyName, secretValue, currentPrefix, currentUpperCase)
							}
							setPrefix(d.Prefix, &currentPrefix)
							setUpper(d.UpperCase, &currentUpperCase)
						}
					} else {
						secretValue, err := vaultClient.ReadSecretKey(c, fmt.Sprintf("%s", f))
						if err != nil {
							return nil, err
						}
						result.Add(fmt.Sprintf("%s", f), secretValue, currentPrefix, currentUpperCase)
					}
				}
				setPrefix(config.Config.Prefix, &currentPrefix)
				setUpper(d.UpperCase, &currentUpperCase)
				setFormat(d.Format, &currentFormat)
			}
		} else {
			secretValue, err := vaultClient.ReadSecret(fmt.Sprintf("%s", a))
			if err != nil {
				return nil, err
			}
			for aa, bb := range secretValue {
				result.Add(aa, bb, currentPrefix, currentUpperCase)
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

func setFormat(potentiaFormat string, currentFormat *string) {
	if potentiaFormat != "" {
		*currentFormat = potentiaFormat
	} else {
		*currentFormat = config.Config.Format
	}
}
