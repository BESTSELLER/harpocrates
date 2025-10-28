package secrets

import (
	"encoding/json"
	"fmt"
	"os"
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

// ToJSON will format a map[string]interface{} to json
func (result Result) ToJSON() string {
	log.Debug().Msg("Exporting as JSON")
	jsonString, err := json.Marshal(result)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to convert result to json")
		os.Exit(1)
	}
	return string(jsonString)
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

func (result Result) ToKVarray(prefix string) (returnString []string) {
	for key, val := range result {
		leKey := fixEnvName(key)
		log.Debug().Msgf("Exporting key: %s", leKey)
		returnString = append(returnString, fmt.Sprintf("%s%s=%v", prefix, leKey, val))
	}
	return returnString
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

// ToYaml exports secrets as yaml
func (result Result) ToYAML() string {
	log.Debug().Msg("Exporting as YAML")
	yamlString, err := yaml.Marshal(result)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to convert result to yaml")
		os.Exit(1)
	}
	return string(yamlString)
}

// fixEnvName replaces all unsported env characters with "_"
func fixEnvName(currentName string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9_]+")
	envVar := reg.ReplaceAllString(currentName, "_")

	return envVar
}

func ToUpperOrNotToUpper(something string, currentUpper *bool) string {
	if *currentUpper {
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
