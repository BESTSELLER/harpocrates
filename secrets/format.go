package secrets

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Result holds the result of all the secrets pulled from Vault
type Result map[string]interface{}

// Add will add a new secret to the Result
func (result Result) Add(key string, value string, prefix string) {
	result[fmt.Sprintf("%s%s", prefix, key)] = value
}

// ToJSON will format a map[string]interface{} to json
func (result Result) ToJSON() string {
	jsonString, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Unable to convert result to json: %s\n", err)
		os.Exit(1)
	}
	return string(jsonString)
}

// ToENV will format a map[string]string to a env file
//
// export KEY='value'
func (result Result) ToENV() string {
	var resturnString string

	for key, val := range result {
		leKey := fixEnvName(key)
		fmt.Println(leKey)
		resturnString += fmt.Sprintf("export %s='%s'\n", strings.ToUpper(leKey), val)
	}
	return resturnString
}

// fixEnvName replaces all unsported env characters with "_"
func fixEnvName(currentName string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9_]+")
	envVar := reg.ReplaceAllString(currentName, "_")

	return envVar
}
