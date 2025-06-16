package vault

import (
	"encoding/json"
	"fmt"
	"strings"
)

const keyNotFound = "the key '%s' was not found in the path '%s': %v"
const secretNotFound = "the secret '%s' was not found: %v"

// ReadSecret from Vault
func (client *API) ReadSecret(path string) (map[string]interface{}, error) {

	secretValues, err := client.Client.Logical().Read(path)
	if err != nil {
		// If Read itself returns an error (e.g., network issue, permission denied before even checking data)
		return nil, fmt.Errorf("failed to read secret from Vault path '%s': %w", path, err)
	}
	if secretValues == nil {
		// This case implies no error from Read, but secretValues is nil (e.g., path does not exist)
		return nil, fmt.Errorf(secretNotFound, path, "path not found or no content")
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
				// Path already contains "data", and we have warnings, likely indicating an issue.
				return nil, fmt.Errorf("error reading secret '%s': %s", path, strings.Join(secretValues.Warnings, ","))
			}

			// Attempt to read from "data/" prefixed path
			appendData := []string{splitPath[0], "data"}
			pathWithData := append(appendData, splitPath[1:]...)
			// Recursive call, error will be propagated as is.
			return client.ReadSecret(strings.Join(pathWithData, "/"))
		}
		// No data and no warnings often means the path is incorrect or doesn't hold data.
		return nil, fmt.Errorf("no data received from secret path '%s'", path)
	}

	b, err := json.Marshal(secretData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal secret data for path '%s': %w", path, err)
	}

	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response from Vault for path '%s': %w", path, err)
	}

	myMap, ok := f.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("secret data from path '%s' is not a map[string]interface{}", path)
	}

	return myMap, nil
}

// ReadSecretKey from Vault
func (client *API) ReadSecretKey(path string, key string) (interface{}, error) {
	secret, err := client.ReadSecret(path)
	// err from ReadSecret already includes path info and should be wrapped if ReadSecret itself failed.
	if err != nil {
		// If ReadSecret failed (e.g. path not found, network error), that error is more relevant.
		// We might not need to wrap it further unless we're adding specific "key not found due to parent error" context.
		// Let's assume ReadSecret's error is sufficient if secret is nil due to such an error.
		return "", fmt.Errorf("could not read secret at path '%s' to find key '%s': %w", path, key, err)
	}
	// If ReadSecret returned a non-nil map but err is also non-nil (shouldn't happen with current ReadSecret logic),
	// or if secret is nil without error (also unlikely), these are edge cases.
	// The primary case here is err == nil and secret != nil.

	secretValue, ok := secret[key]
	if !ok {
		// Key not found in a successfully retrieved secret
		return "", fmt.Errorf(keyNotFound, key, path, "key does not exist in the secret data")
	}

	return secretValue, nil
}
