package files

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"bitbucket.org/bestsellerit/harpocrates/config"
)

// FormatAsJSON will format a map[string]string to json
func FormatAsJSON(input map[string]interface{}) string {
	var result = make(map[string]interface{})

	prefix := getPrefix()
	for key, val := range input {
		leKey := fmt.Sprintf("%s%s", prefix, key)
		result[leKey] = val
	}

	jsonString, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Unable to convert result to json: %s\n", err)
		os.Exit(1)
	}
	return string(jsonString)
}

// FormatAsENV will format a map[string]string to a env file
// export KEY='value'
func FormatAsENV(input map[string]interface{}) string {
	var result string

	prefix := getPrefix()
	for key, val := range input {
		leKey := fmt.Sprintf("%s%s", prefix, key)
		result += fmt.Sprintf("export %s='%s'\n", strings.ToUpper(leKey), val)
	}
	return result
}

func getPrefix() string {
	if config.Config.SecretPrefix != "" {
		return config.Config.SecretPrefix
	}
	return ""
}
