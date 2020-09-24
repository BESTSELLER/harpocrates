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
func (result Result) Add(key string, value interface{}, prefix string, upperCase bool) {
	result[ToUpperOrNotToUpper(fmt.Sprintf("%s%s", prefix, key), &upperCase)] = value
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

func (result Result) toKV(prefix string) string {
	var resturnString string

	for key, val := range result {
		leKey := fixEnvName(key)
		fmt.Println(leKey)
		resturnString += fmt.Sprintf("%s%s='%s'\n", prefix, leKey, val)
	}
	return resturnString
}

// ToENV will format a map[string]string to a env file
//
// export KEY='value'
func (result Result) ToENV() string {
	return result.toKV("export ")
}

// ToK8sSecret exports secrets as raw key values
func (result Result) ToK8sSecret() string {
	return result.toKV("")
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
