package vault

import (
	"encoding/json"
	"fmt"
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
