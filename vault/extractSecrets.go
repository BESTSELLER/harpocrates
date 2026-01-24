package vault

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/adapters/secondary/filesystem"
	vaultadapter "github.com/BESTSELLER/harpocrates/adapters/secondary/vault"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/domain/services"
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

// ExtractSecrets will loop through all those damn interfaces
// This method now delegates to the hexagonal architecture service layer
func (vaultClient *API) ExtractSecrets(input util.SecretJSON, appendToFile bool) ([]Outputs, error) {
	// Create adapters
	vaultAdapter := vaultadapter.NewAdapter(vaultClient.Client)
	filesystemAdapter := filesystem.NewAdapter()
	
	// Create service with adapters
	secretService := services.NewSecretService(vaultAdapter, filesystemAdapter, nil)
	
	// Delegate to service
	serviceOutputs, err := secretService.ExtractSecrets(input, appendToFile)
	if err != nil {
		return nil, err
	}
	
	// Convert service outputs to vault outputs (they have the same structure)
	result := make([]Outputs, len(serviceOutputs))
	for i, out := range serviceOutputs {
		result[i] = Outputs{
			Format:   out.Format,
			Filename: out.Filename,
			Result:   out.Result,
			Owner:    out.Owner,
		}
	}
	
	return result, nil
}

// ExtractSecretsLegacy is the old implementation kept for reference
func (vaultClient *API) ExtractSecretsLegacy(input util.SecretJSON, appendToFile bool) ([]Outputs, error) {
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

					finalResult = append(finalResult, Outputs{Format: currentFormat, Filename: d.FileName, Result: thisResult, Owner: d.Owner})
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
