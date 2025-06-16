package token

import (
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
)

import "fmt"

// Read will read a file and return the content as a string, or an error if reading fails.
func Read() (string, error) {
	// Defaults to Kubernetes Service Account token path
	filePath := "/var/run/secrets/kubernetes.io/serviceaccount/token"

	if config.Config.TokenPath != "" {
		filePath = config.Config.TokenPath
	}

	tokenData, err := files.Read(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read token from '%s': %w", filePath, err)
	}
	return tokenData, nil
}
