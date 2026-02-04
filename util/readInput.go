package util

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/config"
	"go.yaml.in/yaml/v4"
)

// SecretItem represents a union type that can be either a string (simple secret path)
// or a map that will be decoded into a Secret struct (secret with configuration options).
// While the underlying type is 'any' for JSON/YAML unmarshaling compatibility,
// this type alias documents that only string or Secret (as map[string]any) are valid.
type SecretItem = any

// KeyItem represents a union type that can be either a string (simple key name)
// or a map that will be decoded into a SecretKeys struct (key with configuration options).
// While the underlying type is 'any' for JSON/YAML unmarshaling compatibility,
// this type alias documents that only string or SecretKeys (as map[string]any) are valid.
type KeyItem = any

// SecretJSON holds the information about which secrets to fetch and how to save them again
type SecretJSON struct {
	Append        *bool        `json:"append,omitempty"      yaml:"append,omitempty"`
	Format        string       `json:"format,omitempty"      yaml:"format,omitempty"`
	Output        string       `json:"output,omitempty"      yaml:"output,omitempty"`
	Owner         *int         `json:"owner,omitempty"       yaml:"owner,omitempty"`
	Prefix        string       `json:"prefix,omitempty"      yaml:"prefix,omitempty"`
	UpperCase     *bool        `json:"uppercase,omitempty"   yaml:"uppercase,omitempty"`
	Secrets       []SecretItem `json:"secrets,omitempty"     yaml:"secrets,omitempty"`
	GcpWorkloadID bool         `json:"gcpWorkloadID,omitempty"     yaml:"gcpWorkloadID,omitempty"`
}

type Secret struct {
	Append    *bool    `json:"append,omitempty"      yaml:"append,omitempty"`
	Prefix    string   `json:"prefix,omitempty"      yaml:"prefix,omitempty"`
	Format    string   `json:"format,omitempty"      yaml:"format,omitempty"      mapstructure:"format,omitempty"`
	FileName  string   `json:"filename,omitempty"    yaml:"filename,omitempty"    mapstructure:"filename,omitempty"`
	UpperCase *bool    `json:"uppercase,omitempty"   yaml:"uppercase,omitempty"`
	Keys      []KeyItem `json:"keys,omitempty"        yaml:"keys,omitempty"`
	Owner     *int     `json:"owner,omitempty"       yaml:"owner,omitempty"`
}

type SecretKeys struct {
	Append     *bool  `json:"append,omitempty"         yaml:"append,omitempty"`
	Prefix     string `json:"prefix,omitempty"         yaml:"prefix,omitempty"      mapstructure:"prefix,omitempty"`
	UpperCase  *bool  `json:"uppercase,omitempty"      yaml:"uppercase,omitempty"   mapstructure:"uppercase,omitempty"`
	SaveAsFile *bool  `json:"saveAsFile,omitempty"     yaml:"saveAsFile,omitempty"`
}

// IsSecretString checks if a SecretItem is a simple string (secret path)
func IsSecretString(item SecretItem) bool {
	_, ok := item.(string)
	return ok
}

// IsKeyString checks if a KeyItem is a simple string (key name)
func IsKeyString(item KeyItem) bool {
	_, ok := item.(string)
	return ok
}

// GetSecretString safely extracts a string from a SecretItem.
// Returns the string and true if the item is a string, empty string and false otherwise.
func GetSecretString(item SecretItem) (string, bool) {
	s, ok := item.(string)
	return s, ok
}

// GetKeyString safely extracts a string from a KeyItem.
// Returns the string and true if the item is a string, empty string and false otherwise.
func GetKeyString(item KeyItem) (string, bool) {
	s, ok := item.(string)
	return s, ok
}

// GetSecretMap safely extracts a map from a SecretItem for decoding to Secret struct.
// Returns the map and true if the item is a map, nil and false otherwise.
func GetSecretMap(item SecretItem) (map[string]any, bool) {
	m, ok := item.(map[string]any)
	return m, ok
}

// GetKeyMap safely extracts a map from a KeyItem for decoding to SecretKeys struct.
// Returns the map and true if the item is a map, nil and false otherwise.
func GetKeyMap(item KeyItem) (map[string]any, bool) {
	m, ok := item.(map[string]any)
	return m, ok
}

// ReadInput will read the input given to Harpocrates and try to parse it to SecretJSON
// Will also set some default values
func ReadInput(input string) SecretJSON {
	secretJSON := SecretJSON{}
	err := json.Unmarshal([]byte(input), &secretJSON)
	if err == nil {
		goto MoveOn
	}
	err = yaml.Unmarshal([]byte(input), &secretJSON)
	if err != nil {
		fmt.Printf("Your secret file contains an error, please refer to the documentation\n%v\n", err)
		os.Exit(1)
	}

MoveOn:
	if secretJSON.Format != "" {
		config.Config.Format = secretJSON.Format
	}

	if secretJSON.Output == "" {
		secretJSON.Output = "/secrets"
	}
	if config.Config.Output == "" {
		config.Config.Output = secretJSON.Output
	}

	if secretJSON.Owner == nil {
		value := -1
		secretJSON.Owner = &value
	}
	config.Config.Owner = *secretJSON.Owner

	if len(secretJSON.Secrets) == 0 {
		fmt.Println("No secrets provided")
		os.Exit(1)
	}

	config.Config.Prefix = secretJSON.Prefix

	if secretJSON.UpperCase != nil {
		config.Config.UpperCase = *secretJSON.UpperCase
	}

	if secretJSON.Append != nil {
		config.Config.Append = *secretJSON.Append
	}

	if secretJSON.GcpWorkloadID {
		config.Config.GcpWorkloadID = secretJSON.GcpWorkloadID
	}

	return secretJSON
}
