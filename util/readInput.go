package util

import (
	"encoding/json"
	"fmt"
	// "os" // Removed unused import

	"github.com/BESTSELLER/harpocrates/config"
	"gopkg.in/yaml.v3"
)

// SecretJSON holds the information about which secrets to fetch and how to save them again
type SecretJSON struct {
	Append        *bool         `json:"append,omitempty"      yaml:"append,omitempty"`
	Format        string        `json:"format,omitempty"      yaml:"format,omitempty"`
	Output        string        `json:"output,omitempty"      yaml:"output,omitempty"`
	Owner         *int          `json:"owner,omitempty"       yaml:"owner,omitempty"`
	Prefix        string        `json:"prefix,omitempty"      yaml:"prefix,omitempty"`
	UpperCase     *bool         `json:"uppercase,omitempty"   yaml:"uppercase,omitempty"`
	Secrets       []interface{} `json:"secrets,omitempty"     yaml:"secrets,omitempty"`
	GcpWorkloadID bool          `json:"gcpWorkloadID,omitempty"     yaml:"gcpWorkloadID,omitempty"`
}

type Secret struct {
	Append    *bool         `json:"append,omitempty"      yaml:"append,omitempty"`
	Prefix    string        `json:"prefix,omitempty"      yaml:"prefix,omitempty"`
	Format    string        `json:"format,omitempty"      yaml:"format,omitempty"      mapstructure:"format,omitempty"`
	FileName  string        `json:"filename,omitempty"    yaml:"filename,omitempty"    mapstructure:"filename,omitempty"`
	UpperCase *bool         `json:"uppercase,omitempty"   yaml:"uppercase,omitempty"`
	Keys      []interface{} `json:"keys,omitempty"        yaml:"keys,omitempty"`
	Owner     *int          `json:"owner,omitempty"       yaml:"owner,omitempty"`
}

type SecretKeys struct {
	Append     *bool  `json:"append,omitempty"         yaml:"append,omitempty"`
	Prefix     string `json:"prefix,omitempty"         yaml:"prefix,omitempty"      mapstructure:"prefix,omitempty"`
	UpperCase  *bool  `json:"uppercase,omitempty"      yaml:"uppercase,omitempty"   mapstructure:"uppercase,omitempty"`
	SaveAsFile *bool  `json:"saveAsFile,omitempty"     yaml:"saveAsFile,omitempty"`
}

// ReadInput will read the input given to Harpocrates and try to parse it to SecretJSON.
// It also sets some default values and updates the global config.
// TODO: Consider separating config update from parsing.
func ReadInput(input string) (SecretJSON, error) {
	var secretJSON SecretJSON
	var jsonErr, yamlErr error

	// Try unmarshalling as JSON first.
	jsonErr = json.Unmarshal([]byte(input), &secretJSON)
	if jsonErr == nil {
		// Successfully parsed as JSON, proceed to apply defaults and update config.
		return applyDefaultsAndValidate(secretJSON)
	}

	// If JSON failed, try unmarshalling as YAML.
	yamlErr = yaml.Unmarshal([]byte(input), &secretJSON)
	if yamlErr == nil {
		// Successfully parsed as YAML, proceed to apply defaults and update config.
		return applyDefaultsAndValidate(secretJSON)
	}

	// Both JSON and YAML parsing failed.
	// Return a consolidated error message.
	return secretJSON, fmt.Errorf("failed to parse input as JSON or YAML. JSON error: %v, YAML error: %w", jsonErr, yamlErr)
}

// applyDefaultsAndValidate applies default values to the SecretJSON and validates required fields.
// It also updates the global config.Config with values from SecretJSON.
func applyDefaultsAndValidate(sj SecretJSON) (SecretJSON, error) {
	if sj.Format != "" {
		config.Config.Format = sj.Format
	}

	if sj.Output == "" {
		sj.Output = "/secrets" // Default output path
	}
	config.Config.Output = sj.Output

	if sj.Owner == nil {
		defaultValue := -1
		sj.Owner = &defaultValue
	}
	config.Config.Owner = *sj.Owner

	if len(sj.Secrets) == 0 {
		return sj, fmt.Errorf("no secrets provided in the input")
	}

	config.Config.Prefix = sj.Prefix

	if sj.UpperCase != nil {
		config.Config.UpperCase = *sj.UpperCase
	}

	if sj.Append != nil {
		config.Config.Append = *sj.Append
	}

	if sj.GcpWorkloadID {
		config.Config.GcpWorkloadID = sj.GcpWorkloadID
	}

	return sj, nil
}
