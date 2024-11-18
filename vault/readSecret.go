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

	b, err := json.Marshal(secretData)
	if err != nil {
		return nil, err
	}

	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response from Vault")
	}

	myMap := f.(map[string]interface{})

	return myMap, nil
}

// ReadSecretKey from Vault
func (client *API) ReadSecretKey(path string, key string) (interface{}, error) {
	secret, err := client.ReadSecret(path)
	if secret == nil {
		return "", fmt.Errorf(keyNotFound, key, path, err)
	}
	if err != nil {
		return "", err
	}
	secretKey := secret[key]
	if secretKey == nil {
		return "", fmt.Errorf(keyNotFound, key, path, err)
	}

	return secretKey, nil
}
