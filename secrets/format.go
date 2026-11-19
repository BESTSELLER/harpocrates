package secrets

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"go.yaml.in/yaml/v4"
)

// Result holds the result of all the secrets pulled from Vault
type Result map[string]any

// envNameRegexp is a precompiled regular expression used by fixEnvName
var envNameRegexp = regexp.MustCompile("[^a-zA-Z0-9_]+")

// Add will add a new secret to the Result
func (result Result) Add(key string, value any, prefix string, upperCase bool) {
	result[ToUpperOrNotToUpper(fmt.Sprintf("%s%s", prefix, key), &upperCase)] = value
}

// ToJSON will format a map[string]any to json
func (result Result) ToJSON() string {
	log.Debug().Msg("Exporting as JSON")
	jsonString, err := json.Marshal(result)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to convert result to json")
	}
	return string(jsonString)
}

func (result Result) toKV(prefix string) string {
	var returnString strings.Builder

	for key, val := range result {
		leKey := fixEnvName(key)
		log.Info().Msgf("Exporting key: %s", leKey)
		fmt.Fprintf(&returnString, "%s%s=%s\n", prefix, leKey, getStringRepresentation(val))
	}
	return returnString.String()
}

// ToKVarray converts the result to a key=value array
func (result Result) ToKVarray(prefix string) (returnString []string) {
	for key, val := range result {
		leKey := fixEnvName(key)
		log.Debug().Msgf("Exporting key: %s", leKey)
		returnString = append(returnString, fmt.Sprintf("%s%s=%v", prefix, leKey, val))
	}
	return returnString
}

func (result Result) toSecretKV() string {
	var returnString strings.Builder

	for key, val := range result {
		log.Info().Msgf("Exporting key: %s", key)
		fmt.Fprintf(&returnString, "%s=%s\n", key, getStringRepresentation(val))
	}
	return returnString.String()
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

// ToYAML exports secrets as yaml
func (result Result) ToYAML() string {
	log.Debug().Msg("Exporting as YAML")
	yamlString, err := yaml.Marshal(result)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to convert result to yaml")
	}
	return string(yamlString)
}

// ToINI exports secrets as an INI file. Keys holding objects become sections,
// top-level scalars go into [default]. Sections and keys are sorted.
func (result Result) ToINI() string {
	log.Debug().Msg("Exporting as INI")

	sections := map[string]map[string]string{}

	// add fails loudly on sanitisation collisions: map order is random, so merging
	// would silently lose a value nondeterministically between runs.
	add := func(section, key, value string) {
		if existing, ok := sections[section][key]; ok && existing != value {
			log.Fatal().Msgf("INI key collision in section [%s] after sanitisation: %q", section, key)
		}
		sections[section][key] = value
	}

	for key, val := range result {
		if nested, ok := val.(map[string]any); ok {
			sectionName := fixEnvName(key)
			if sections[sectionName] == nil {
				sections[sectionName] = map[string]string{}
			}
			for nestedKey, nestedVal := range nested {
				add(sectionName, fixEnvName(nestedKey), toINIValue(nestedVal))
			}
			continue
		}
		if _, isDefaultSection := result["default"].(map[string]any); isDefaultSection {
			log.Warn().Msg("A key named 'default' merges with the synthetic [default] section")
		}
		if sections["default"] == nil {
			sections["default"] = map[string]string{}
		}
		add("default", fixEnvName(key), toINIValue(val))
	}

	var sectionNames []string
	for name := range sections {
		sectionNames = append(sectionNames, name)
	}
	sort.Strings(sectionNames)

	var blocks []string
	for _, name := range sectionNames {
		keys := make([]string, 0, len(sections[name]))
		for key := range sections[name] {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		var block strings.Builder
		fmt.Fprintf(&block, "[%s]\n", name)
		for _, key := range keys {
			fmt.Fprintf(&block, "%s = %s\n", key, sections[name][key])
		}
		blocks = append(blocks, block.String())
	}

	return strings.Join(blocks, "\n")
}

// toINIValue renders a value raw (no quotes) so INI readers see the literal value
func toINIValue(val any) string {
	switch v := val.(type) {
	case nil:
		return "null"
	case string:
		return strings.ReplaceAll(strings.ReplaceAll(v, "\n", `\n`), "\r", `\r`)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// fixEnvName replaces all unsupported env characters with "_"
func fixEnvName(currentName string) string {
	envVar := envNameRegexp.ReplaceAllString(currentName, "_")

	return envVar
}

// ToUpperOrNotToUpper changes the case of a string depending on the value of currentUpper
func ToUpperOrNotToUpper(something string, currentUpper *bool) string {
	if *currentUpper {
		return strings.ToUpper(something)
	}
	return something
}

func getStringRepresentation(val any) string {
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
