package vault

import (
	"encoding/json"
	"fmt"
	"log"
)

const keyNotFound = "The key '%s' was not found in the path '%s'\n"
const secretNotFound = "The secret '%s' was not found \n"

// ReadSecret from Vault
func (client *API) ReadSecret(path string) map[string]interface{} {

	secretValues, err := client.Client.Logical().Read(path)
	if err != nil {
		log.Fatalln(err)
	}

	if secretValues == nil {
		log.Fatalf(secretNotFound, path)
	}

	secretData := secretValues.Data["data"]

	if secretData == nil {
		fmt.Println("secretValues", secretValues)
		secretData = secretValues.Data
	}

	b, err := json.Marshal(secretData)
	if err != nil {
		log.Fatalln(err)
	}

	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		log.Fatalln("Unable to unmarshal response from Vault")
	}

	myMap := f.(map[string]interface{})

	return myMap
}

// ReadSecretKey from Vault
func (client *API) ReadSecretKey(path string, key string) string {
	secret := client.ReadSecret(path)
	if secret == nil {
		log.Fatalf(keyNotFound, key, path)
	}
	secretKey := secret[key]
	if secretKey == nil {
		log.Fatalf(keyNotFound, key, path)
	}

	return secretKey.(string)
}
