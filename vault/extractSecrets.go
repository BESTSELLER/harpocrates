package vault

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/adapters/secondary/filesystem"
	vaultadapter "github.com/BESTSELLER/harpocrates/adapters/secondary/vault"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/domain/models"
	"github.com/BESTSELLER/harpocrates/domain/services"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/secrets"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/go-viper/mapstructure/v2"
)

// ExtractSecrets will loop through all those damn interfaces
// This method now delegates to the hexagonal architecture service layer
func (vaultClient *API) ExtractSecrets(input util.SecretJSON, appendToFile bool) ([]models.SecretOutput, error) {
	// Create adapters
	vaultAdapter := vaultadapter.NewAdapter(vaultClient.Client)
	filesystemAdapter := filesystem.NewAdapter()
	
	// Create service with adapters
	secretService := services.NewSecretService(vaultAdapter, filesystemAdapter, nil)
	
	// Delegate to service - now returns models.SecretOutput directly
	return secretService.ExtractSecrets(input, appendToFile)
}

// ExtractSecretsLegacy is the old implementation kept for reference
func (vaultClient *API) ExtractSecretsLegacy(input util.SecretJSON, appendToFile bool) ([]models.SecretOutput, error) {
	var finalResult []models.SecretOutput
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
			mapstructure.Decode(b, &aa)

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

					finalResult = append(finalResult, models.SecretOutput{Format: currentFormat, Filename: d.FileName, Result: thisResult, Owner: d.Owner})
					continue
				}

				for _, f := range d.Keys {
					// If the key is just a secret path, then it will read that from Vault, otherwise:
					if fmt.Sprintf("%T", f) != "string" {
						bb := map[string]util.SecretKeys{}
						mapstructure.Decode(f, &bb)

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

	finalResult = append(finalResult, models.SecretOutput{Format: config.Config.Format, Filename: "", Result: result})
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
