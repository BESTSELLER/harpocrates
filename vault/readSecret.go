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

// ReadSecretKey retrieves a value from a Vault secret at a specific path.
//
// It supports accessing nested keys using dot notation or array brackets.
//
// Usage examples (assuming JSON structure):
//   - "simpleKey"           -> {"simpleKey": "value"}
//   - "nested.key"          -> {"nested": {"key": "value"}}
//   - "array[0]"            -> {"array": ["value", "other"]}
//   - "mixed.array[0].key"  -> {"mixed": {"array": [{"key": "value"}]}}
//   - "key.with.dots"       -> {"key": {"with.dots": "value"}}
//   - "key.with.dots.child" -> {"key": {"with.dots": {"child": "value"}}}
func (client *API) ReadSecretKey(path string, secretKey string) (interface{}, error) {
	secret, err := client.ReadSecret(path)
	if secret == nil {
		return "", fmt.Errorf(keyNotFound, secretKey, path, err)
	}
	if err != nil {
		return "", err
	}

	// 1. Literal match
	// Check if the secretKey exists exactly as-is in the top-level secret.
	// This handles keys that naturally contain dots or brackets without needing traversal.
	if literalValue, keyExists := secret[secretKey]; keyExists {
		return literalValue, nil
	}

	// 2. Traversal configuration
	// Normalize access syntax by replacing array brackets with dots to unify the traversal loop.
	// e.g., "users[0].name" becomes "users.0.name"
	normalizedKey := strings.ReplaceAll(secretKey, "[", ".")
	normalizedKey = strings.ReplaceAll(normalizedKey, "]", "")
	keys := strings.Split(normalizedKey, ".")

	var current interface{} = secret
	for i := 0; i < len(keys); i++ {
		keySegment := keys[i]

		if currentMap, isMap := current.(map[string]interface{}); isMap {
			// Check for exact match of the current segment in the map
			if mapValue, exists := currentMap[keySegment]; exists {
				current = mapValue
				continue
			}

			// Attempt to match keys with dots (merging segments)
			// This handles cases where a JSON key contains dots (e.g., "labels.app")
			// but isn't necessarily at the end of the path.
			matchFound := false
			for j := i + 1; j < len(keys); j++ {
				// Construct candidate key from segments i to j (inclusive)
				candidate := strings.Join(keys[i:j+1], ".")
				if mapValue, exists := currentMap[candidate]; exists {
					current = mapValue
					i = j // Advance the outer loop index
					matchFound = true
					break
				}
			}

			if matchFound {
				continue
			}
		}

		if currentSlice, isSlice := current.([]interface{}); isSlice {
			// Handle array index access (e.g., from "users[0]" -> "0")
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
