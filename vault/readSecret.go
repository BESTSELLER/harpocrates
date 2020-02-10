package vault

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadSecret from Vault
func ReadSecret(path string) map[string]interface{} {
	client := createClient()

	secretValues, err := client.Logical().Read(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// if secretValues.Data == nil {
	// 	secretData := secretValues
	// } else {

	// }

	secretData := secretValues.Data["data"]

	if secretData == nil {
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
	secretKey := secret[key]
	if secretKey == nil {
		fmt.Printf("The key '%s' was not found in the path '%s'\n", key, path)
		os.Exit(1)
	}

	return secretKey.(string)
}
