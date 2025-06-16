package secrets

import (
	"encoding/json"
	"fmt"
	// "os" // Removed unused import
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Result holds the result of all the secrets pulled from Vault
type Result map[string]interface{}

// Add will add a new secret to the Result
func (result Result) Add(key string, value interface{}, prefix string, upperCase bool) {
	result[ToUpperOrNotToUpper(fmt.Sprintf("%s%s", prefix, key), &upperCase)] = value
}

// ToJSON will format a map[string]interface{} to json.
// It returns the JSON string or an error if marshalling fails.
func (result Result) ToJSON() (string, error) {
	log.Debug().Msg("Exporting as JSON")
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("unable to convert result to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

func (result Result) toKV(prefix string) string {
	var resturnString string

	for key, val := range result {
		leKey := fixEnvName(key)
		log.Info().Msgf("Exporting key: %s", leKey)
		resturnString += fmt.Sprintf("%s%s=%s\n", prefix, leKey, getStringRepresentation(val))
	}
	return resturnString
}

func (result Result) toSecretKV() string {
	var resturnString string

	for key, val := range result {
		log.Info().Msgf("Exporting key: %s", key)
		resturnString += fmt.Sprintf("%s=%s\n", key, getStringRepresentation(val))
	}
	return resturnString
}

// ToENV will format a map[string]string to a env file
//
// export KEY='value'
func (result Result) ToENV() string {
	log.Debug().Msg("Exporting as env values")
	return result.toKV("export ")
}

// ToK8sSecret exports secrets as raw key values
func (result Result) ToK8sSecret() string {
	log.Debug().Msg("Exporting as raw key values")
	return result.toSecretKV()
}

// ToYAML exports secrets as yaml.
// It returns the YAML string or an error if marshalling fails.
func (result Result) ToYAML() (string, error) {
	log.Debug().Msg("Exporting as YAML")
	yamlBytes, err := yaml.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("unable to convert result to YAML: %w", err)
	}
	return string(yamlBytes), nil
}

// fixEnvName replaces all unsported env characters with "_"
func fixEnvName(currentName string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9_]+")
	envVar := reg.ReplaceAllString(currentName, "_")

	return envVar
}

func ToUpperOrNotToUpper(something string, currentUpper *bool) string {
	// Check if the pointer is nil before dereferencing
	if currentUpper != nil && *currentUpper {
		return strings.ToUpper(something)
	}
	return something
}

func getStringRepresentation(val interface{}) string {
	switch val.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case int:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%f", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("'%s'", val)
	}
}
