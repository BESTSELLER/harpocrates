package services

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/domain/ports"
	"github.com/BESTSELLER/harpocrates/secrets"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/go-viper/mapstructure/v2"
)

// SecretService orchestrates secret extraction and processing
type SecretService struct {
	fetcher       ports.SecretFetcher
	writer        ports.SecretWriter
	authenticator ports.Authenticator
}

// NewSecretService creates a new instance of SecretService
func NewSecretService(fetcher ports.SecretFetcher, writer ports.SecretWriter, authenticator ports.Authenticator) *SecretService {
	return &SecretService{
		fetcher:       fetcher,
		writer:        writer,
		authenticator: authenticator,
	}
}

// Outputs represents the output format for secrets
type Outputs struct {
	Format   string         `json:"format,omitempty"    yaml:"format,omitempty"`
	Filename string         `json:"filename,omitempty"  yaml:"filename,omitempty"`
	Result   secrets.Result `json:"result,omitempty"    yaml:"result,omitempty"`
	Owner    *int           `json:"owner,omitempty"     yaml:"owner,omitempty"`
}

// ExtractSecrets extracts secrets based on the input configuration
func (s *SecretService) ExtractSecrets(input util.SecretJSON, appendToFile bool) ([]Outputs, error) {
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
				s.setPrefix(d.Prefix, &currentPrefix)
				s.setUpper(d.UpperCase, &currentUpperCase)
				s.setFormat(d.Format, &currentFormat)

				if len(d.Keys) == 0 {
					secretValue, err := s.fetcher.ReadSecret(c)
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
							s.setPrefix(i.Prefix, &currentPrefix)
							s.setUpper(d.UpperCase, &currentUpperCase)

							if i.SaveAsFile != nil {
								secretValue, err := s.fetcher.ReadSecretKey(c, h)
								if err != nil {
									return nil, err
								}
								if *i.SaveAsFile {
									// Write directly using writer port
									err = s.writer.Write(input.Output, secrets.ToUpperOrNotToUpper(fmt.Sprintf("%s%s", currentPrefix, h), &currentUpperCase), secretValue, nil, appendToFile)
									if err != nil {
										return nil, err
									}
								} else {
									result.Add(h, secretValue, currentPrefix, currentUpperCase)
								}
							} else {
								secretValue, err := s.fetcher.ReadSecretKey(c, h)
								if err != nil {
									return nil, err
								}
								result.Add(h, secretValue, currentPrefix, currentUpperCase)
							}
							s.setPrefix(d.Prefix, &currentPrefix)
							s.setUpper(d.UpperCase, &currentUpperCase)
						}
					} else {
						secretValue, err := s.fetcher.ReadSecretKey(c, fmt.Sprintf("%s", f))
						if err != nil {
							return nil, err
						}
						result.Add(fmt.Sprintf("%s", f), secretValue, currentPrefix, currentUpperCase)
					}
				}
				s.setPrefix(config.Config.Prefix, &currentPrefix)
				s.setUpper(d.UpperCase, &currentUpperCase)
				s.setFormat(d.Format, &currentFormat)
			}
		} else {
			secretValue, err := s.fetcher.ReadSecret(fmt.Sprintf("%s", a))
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

func (s *SecretService) setPrefix(potentialPrefix string, currentPrefix *string) {
	if potentialPrefix != "" {
		*currentPrefix = potentialPrefix
	} else {
		*currentPrefix = config.Config.Prefix
	}
}

func (s *SecretService) setUpper(potentialUpper *bool, currentUpper *bool) {
	if potentialUpper != nil {
		*currentUpper = *potentialUpper
	} else {
		*currentUpper = config.Config.UpperCase
	}
}

func (s *SecretService) setFormat(potentiaFormat string, currentFormat *string) {
	if potentiaFormat != "" {
		*currentFormat = potentiaFormat
	} else {
		*currentFormat = config.Config.Format
	}
}
