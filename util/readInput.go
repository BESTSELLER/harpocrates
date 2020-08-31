package util

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/config"
)

// SecretJSON holds the information about which secrets to fetch and how to save them again
type SecretJSON struct {
	Format  string        `json:"format,omitempty"`
	DirPath string        `json:"dirPath,omitempty"`
	Prefix  string        `json:"prefix,omitempty"`
	Secrets []interface{} `json:"secrets,omitempty"`
}

// ReadInput will read the input given to Harpocrates and try to parse it to SecretJSON
// Will also set some default values
func ReadInput(input string) SecretJSON {
	secretJSON := SecretJSON{}
	err := json.Unmarshal([]byte(input), &secretJSON)
	if err != nil {
		fmt.Printf("Your secret file contains an error, please refer to the documentation: %v\n", err)
		os.Exit(1)
	}
	if secretJSON.Format == "" {
		secretJSON.Format = "json"
	} else {
		if secretJSON.Format != "json" && secretJSON.Format != "env" {
			fmt.Println("An invalid format was provided, only these formats are allowed at the moment:\njson\nenv")
			os.Exit(1)
		}
	}

	if secretJSON.DirPath == "" {
		secretJSON.DirPath = "/secrets"
	}
	if len(secretJSON.Secrets) == 0 {
		fmt.Println("No secrets provided")
		os.Exit(1)
	}

	config.Config.SecretPrefix = secretJSON.Prefix

	return secretJSON
}
