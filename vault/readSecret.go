package vault

import (
	"encoding/json"
	"fmt"
	"os"
)

const keyNotFound = "The key '%s' was not found in the path '%s'\n"
const secretNotFound = "The secret '%s' was not found \n"

// ReadSecret from Vault
func ReadSecret(path string) map[string]interface{} {
	client := createClient()

	secretValues, err := client.Logical().Read(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if secretValues == nil {
		fmt.Printf(secretNotFound, path)
		os.Exit(1)
	}

	secretData := secretValues.Data["data"]

	if secretData == nil {
		fmt.Println("secretValues", secretValues)
		secretData = secretValues.Data
	}

	b, err := json.Marshal(secretData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		fmt.Println("Unable to unmarshal response from Vault")
		os.Exit(1)
	}

	myMap := f.(map[string]interface{})

	return myMap
}

// ReadSecretKey from Vault
func ReadSecretKey(path string, key string) string {
	secret := ReadSecret(path)
	if secret == nil {
		fmt.Printf(keyNotFound, key, path)
		os.Exit(1)
	}
	secretKey := secret[key]
	if secretKey == nil {
		fmt.Printf(keyNotFound, key, path)
		os.Exit(1)
	}

	return secretKey.(string)
}
