package files

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/BESTSELLER/harpocrates/config"
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
		leKey := fixEnvName(fmt.Sprintf("%s%s", prefix, key))
		fmt.Println(leKey)
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

// fixEnvName replaces all unsported env characters with "_"
func fixEnvName(currentName string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9_]+")
	envVar := reg.ReplaceAllString(currentName, "_")

	return envVar
}
