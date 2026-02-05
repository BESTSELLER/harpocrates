package vault

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const keyNotFound = "the key '%s' was not found in the path '%s': %v"
const secretNotFound = "the secret '%s' was not found: %v"

// ReadSecret from Vault
func (client *API) ReadSecret(path string) (map[string]interface{}, error) {

	secretValues, err := client.Client.Logical().Read(path)
	if secretValues == nil {
		return nil, fmt.Errorf(secretNotFound, path, err)
	}

	secretData := secretValues.Data["data"]

	if secretData == nil {
		secretData = secretValues.Data
	}

	// Will append data to path and retry if "data" is empty and warnings is present
	// if path contains data and warnings an error is returned
	if fmt.Sprintf("%s", secretValues.Data) == fmt.Sprintf("%s", make(map[string]interface{})) {
		if len(secretValues.Warnings) > 0 {
			splitPath := strings.Split(path, "/")
			if splitPath[1] == "data" {
				return nil, fmt.Errorf("%s", strings.Join(secretValues.Warnings, ","))
			}

			appendData := []string{splitPath[0], "data"}
			pathWithData := append(appendData, splitPath[1:]...)
			return client.ReadSecret(strings.Join(pathWithData, "/"))

		}
		return nil, fmt.Errorf("no data recieved")
	}

	jsonBytes, err := json.Marshal(secretData)
	if err != nil {
		return nil, err
	}

	var secretInterface interface{}
	err = json.Unmarshal(jsonBytes, &secretInterface)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response from Vault")
	}

	secretMap := secretInterface.(map[string]interface{})

	return secretMap, nil
}

// ReadSecretKey from Vault
func (client *API) ReadSecretKey(path string, secretKey string) (interface{}, error) {
	secret, err := client.ReadSecret(path)
	if secret == nil {
		return "", fmt.Errorf(keyNotFound, secretKey, path, err)
	}
	if err != nil {
		return "", err
	}

	// 1. Literal match
	if literalValue, keyExists := secret[secretKey]; keyExists {
		return literalValue, nil
	}

	// 2. Traversal
	normalizedKey := strings.ReplaceAll(secretKey, "[", ".")
	normalizedKey = strings.ReplaceAll(normalizedKey, "]", "")
	keys := strings.Split(normalizedKey, ".")

	var current interface{} = secret
	for i := 0; i < len(keys); i++ {
		keySegment := keys[i]

		if currentMap, isMap := current.(map[string]interface{}); isMap {
			if mapValue, exists := currentMap[keySegment]; exists {
				current = mapValue
				continue
			}

			// Attempt to match keys with dots (merging segments)
			mergedKeys := strings.Join(keys[i:], ".")
			if mapValue, exists := currentMap[mergedKeys]; exists {
				current = mapValue
				break
			}
		}

		if currentSlice, isSlice := current.([]interface{}); isSlice {
			if sliceIndex, err := strconv.Atoi(keySegment); err == nil {
				if sliceIndex >= 0 && sliceIndex < len(currentSlice) {
					current = currentSlice[sliceIndex]
					continue
				}
			}
		}
		return "", fmt.Errorf(keyNotFound, secretKey, path, nil)
	}
	return current, nil
}
