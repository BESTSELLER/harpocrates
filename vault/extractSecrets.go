package vault

import (
	"fmt"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/secrets"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/go-viper/mapstructure/v2"
	"github.com/rs/zerolog/log" // Added import for log
)

type Outputs struct {
	Format   string         `json:"format,omitempty"    yaml:"format,omitempty"`
	Filename string         `json:"filename,omitempty"  yaml:"filename,omitempty"`
	Result   secrets.Result `json:"result,omitempty"    yaml:"result,omitempty"`
	Owner    *int           `json:"owner,omitempty"     yaml:"owner,omitempty"`
}

// secretContext holds the processing state (prefix, uppercase, format) for a secret.
type secretContext struct {
	prefix    string
	upperCase bool
	format    string
}

// newSecretContext creates a new secretContext with default values from global config.
func newSecretContext() secretContext {
	return secretContext{
		prefix:    config.Config.Prefix,
		upperCase: config.Config.UpperCase,
		format:    config.Config.Format,
	}
}

// update updates the context with values from a util.Secret, util.SecretKeys, or resets to default.
func (sc *secretContext) update(prefixOverride string, upperCaseOverride *bool, formatOverride string, resetToGlobalDefaults bool) secretContext {
	newCtx := *sc // Create a copy.

	if prefixOverride != "" {
		newCtx.prefix = prefixOverride
	} else if resetToGlobalDefaults {
		newCtx.prefix = config.Config.Prefix // Reset to global if explicitly asked
	}
	// If prefixOverride is empty and resetToGlobalDefaults is false, newCtx.prefix retains sc.prefix (inherited)

	if upperCaseOverride != nil {
		newCtx.upperCase = *upperCaseOverride
	} else if resetToGlobalDefaults {
		newCtx.upperCase = config.Config.UpperCase
	}
	// Inherits sc.upperCase if override is nil and not resetting

	if formatOverride != "" {
		newCtx.format = formatOverride
	} else if resetToGlobalDefaults {
		newCtx.format = config.Config.Format
	}
	// Inherits sc.format if override is empty and not resetting

	return newCtx
}

// processSingleSecretEntry processes a single secret entry, which can be a string or a map.
func (vaultClient *API) processSingleSecretEntry(entry interface{}, appendToFile bool, baseContext secretContext) ([]Outputs, secrets.Result, error) {
	var entryOutputs []Outputs
	var entryResult = make(secrets.Result)

	switch E := entry.(type) {
	case string: // Simple secret path
		secretValue, err := vaultClient.ReadSecret(E)
		if err != nil {
			return nil, nil, err
		}
		for k, v := range secretValue {
			entryResult.Add(k, v, baseContext.prefix, baseContext.upperCase)
		}
	case map[string]interface{}: // Complex secret configuration
		var complexSecrets map[string]util.Secret
		if err := mapstructure.Decode(E, &complexSecrets); err != nil {
			return nil, nil, fmt.Errorf("failed to decode complex secret config: %w", err)
		}

		for secretPath, secretConfig := range complexSecrets {
			// For a new secret path, overrides apply. If an override is empty/nil, it means reset to global default for that specific setting.
			currentPathContext := baseContext.update(secretConfig.Prefix, secretConfig.UpperCase, secretConfig.Format, true)

			if len(secretConfig.Keys) == 0 { // Process as a full secret
				secretValue, err := vaultClient.ReadSecret(secretPath)
				if err != nil {
					return nil, nil, err
				}
				var pathResult = make(secrets.Result)
				for k, v := range secretValue {
					pathResult.Add(k, v, currentPathContext.prefix, currentPathContext.upperCase)
				}
				entryOutputs = append(entryOutputs, Outputs{
					Format:   currentPathContext.format,
					Filename: secretConfig.FileName,
					Result:   pathResult,
					Owner:    secretConfig.Owner,
				})
			} else { // Process specific keys within the secret
				// The result for keys under a specific path might be aggregated or handled per key.
				// Current logic seems to aggregate them into `entryResult` unless `SaveAsFile` is true.
				keyResults, err := vaultClient.processSecretKeys(secretPath, secretConfig.Keys, appendToFile, currentPathContext, secretConfig.FileName, secretConfig.Owner)
				if err != nil {
					return nil, nil, err
				}
				// If keys generate direct outputs (e.g., SaveAsFile), they are in keyResults.
				// If they modify a shared result map, ensure it's handled correctly.
				// This part needs careful review based on expected aggregation.
				// Assuming processSecretKeys returns outputs for SaveAsFile and updates a result map for others.
				// For now, let's assume processSecretKeys returns individual file outputs and the main result map is built up.
				for _, ko := range keyResults {
					if ko.Filename != "" { // This indicates a SaveAsFile scenario output
						entryOutputs = append(entryOutputs, ko)
					} else {
						// keyOutput.Result (ko.Result) contains keys that are already correctly prefixed and cased
						// by processSecretKeys using its specific keyCtx.
						// Merge them directly into entryResult for this secret path.
						for k, v := range ko.Result {
							entryResult[k] = v // Direct assignment, no further prefixing or case change here
						}
					}
				}
			}
		}
	default:
		return nil, nil, fmt.Errorf("unexpected type in secrets array: %T", entry)
	}
	return entryOutputs, entryResult, nil
}

// processSecretKeys processes individual keys within a complex secret configuration.
func (vaultClient *API) processSecretKeys(secretPath string, keys []interface{}, appendToFile bool, parentContext secretContext, defaultFilename string, defaultOwner *int) ([]Outputs, error) {
	var keyOutputs []Outputs
	// Results for keys not saved as files are aggregated by the caller, or this function needs to return them.
	// Let's clarify: if a key is NOT SaveAsFile, its result is added to the parent's result map.
	// This function will return outputs for SaveAsFile, and the caller will handle other results.

	for _, keyEntry := range keys {
		switch K := keyEntry.(type) {
		case string: // Simple key name
			secretValue, err := vaultClient.ReadSecretKey(secretPath, K)
			if err != nil {
				return nil, err
			}
			// This result needs to be passed back to be added to the parent's result map.
			// For now, creating a temporary output. This needs to be handled by the caller.
			tempResult := make(secrets.Result)
			tempResult.Add(K, secretValue, parentContext.prefix, parentContext.upperCase)
			keyOutputs = append(keyOutputs, Outputs{Result: tempResult, Format: parentContext.format}) // No filename means add to main result
		case map[string]interface{}: // Key with specific configuration (e.g., SaveAsFile)
			var keyConfigMap map[string]util.SecretKeys
			if err := mapstructure.Decode(K, &keyConfigMap); err != nil {
				return nil, fmt.Errorf("failed to decode key config: %w", err)
			}

			for keyName, keySettings := range keyConfigMap {
				// Key-specific context overrides. If an override is empty/nil, inherit from parentContext.
				// Do not reset to global defaults unless explicitly specified by keySettings (which isn't current model).
				keyCtx := parentContext.update(keySettings.Prefix, keySettings.UpperCase, "", false)

				secretValue, err := vaultClient.ReadSecretKey(secretPath, keyName)
				if err != nil {
					return nil, err
				}

				if keySettings.SaveAsFile != nil && *keySettings.SaveAsFile {
					// Determine filename: keyName takes precedence if defaultFilename is empty or not applicable.
					// The logic for `secrets.ToUpperOrNotToUpper` implies filename might also be subject to case changes.
					// Assuming filename here means the actual file path, not just the key.

					// DIAGNOSTIC CODE REMOVED - Using keyCtx.prefix directly
					outputFileName := secrets.ToUpperOrNotToUpper(fmt.Sprintf("%s%s", keyCtx.prefix, keyName), &keyCtx.upperCase)
					// files.Write needs output directory, which is global (config.Config.Output)
					// Owner for SaveAsFile seems to be from the parent secretConfig (defaultOwner)
					files.Write(config.Config.Output, outputFileName, secretValue, defaultOwner, appendToFile)
					log.Debug().Msgf("Secret key '%s' written to file: %s/%s", keyName, config.Config.Output, outputFileName)
					// Create an Output entry to signify this action, though it's directly written.
					// This might be useful for logging or tracking.
					keyOutputs = append(keyOutputs, Outputs{
						Filename: outputFileName,       // Reflects the actual filename used
						Format:   parentContext.format, // Or a specific format if SaveAsFile implies one
						// Result might be empty or contain reference, as it's written to file
					})
				} else {
					// Add to parent's result map - signal this by returning a result in Outputs
					tempResult := make(secrets.Result)

					// DIAGNOSTIC CODE REMOVED - Using keyCtx.prefix directly
					tempResult.Add(keyName, secretValue, keyCtx.prefix, keyCtx.upperCase)
					keyOutputs = append(keyOutputs, Outputs{Result: tempResult, Format: parentContext.format})
				}
			}
		default:
			return nil, fmt.Errorf("unexpected type for key entry: %T", keyEntry)
		}
	}
	return keyOutputs, nil
}

// ExtractSecrets loops through secret configurations and extracts them from Vault.
func (vaultClient *API) ExtractSecrets(input util.SecretJSON, appendToFile bool) ([]Outputs, error) {
	var allOutputs []Outputs
	var aggregatedResult = make(secrets.Result) // For secrets not written to individual files

	baseContext := newSecretContext()

	for _, entry := range input.Secrets {
		outputs, result, err := vaultClient.processSingleSecretEntry(entry, appendToFile, baseContext)
		if err != nil {
			return nil, fmt.Errorf("error processing secret entry '%v': %w", entry, err)
		}
		allOutputs = append(allOutputs, outputs...)
		// The 'result' (renamed to resultFromEntry for clarity) already has its keys correctly prefixed
		// and case-transformed by processSingleSecretEntry based on its specific context.
		// So, just merge them directly into aggregatedResult.
		for k, v := range result {
			aggregatedResult[k] = v // Direct assignment, no further prefixing or case change
		}
	}

	// Add the aggregated result as a final output, using global format settings
	// This assumes that if specific files are written, their outputs are already in allOutputs.
	// If aggregatedResult is not empty, it means some secrets were meant for the default output file.
	if len(aggregatedResult) > 0 {
		allOutputs = append(allOutputs, Outputs{
			Format:   config.Config.Format, // Global format for the main aggregated output
			Filename: "",                   // Empty filename indicates default output behavior
			Result:   aggregatedResult,
			// Owner for aggregated result? Uses global default implicitly if files.Write is called later.
		})
	}

	return allOutputs, nil
}
