package vault

import (
	"encoding/json"
	"fmt"
	"strings"
)

const keyNotFound = "The key '%s' was not found in the path '%s'\n"
const secretNotFound = "The secret '%s' was not found \n"

// ReadSecret from Vault
func (client *API) ReadSecret(path string) (map[string]interface{}, error) {

	secretValues, _ := client.Client.Logical().Read(path)
	if secretValues == nil {
		return nil, fmt.Errorf(secretNotFound, path)
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
		return nil, fmt.Errorf("No data recieved")
	}

	b, err := json.Marshal(secretData)
	if err != nil {
		return nil, err
	}

	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response from Vault")
	}

	myMap := f.(map[string]interface{})

	return myMap, nil
}

// ReadSecretKey from Vault
func (client *API) ReadSecretKey(path string, key string) (string, error) {
	secret, err := client.ReadSecret(path)
	if secret == nil {
		return "", fmt.Errorf(keyNotFound, key, path)
	}
	if err != nil {
		return "", err
	}
	secretKey := secret[key]
	if secretKey == nil {
		return "", fmt.Errorf(keyNotFound, key, path)
	}

	return secretKey.(string), nil
}
