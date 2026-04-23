package token

import (
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
)

// Read will read a file and return the content as string
func Read() (string, error) {
	// Defaults to Kubernetes Service Account
	filePath := "/var/run/secrets/kubernetes.io/serviceaccount/token"

	if config.Config.TokenPath != "" {
		filePath = config.Config.TokenPath
	}

	return files.Read(filePath)
}
