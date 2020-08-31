package files

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BESTSELLER/harpocrates/config"
)

// ReadTokenFile will read a file and return the content as string
func ReadTokenFile() string {
	// Defaults to Kubernetes Service Account
	filePath := "/var/run/secrets/kubernetes.io/serviceaccount/token"

	if config.Config.TokenPath != "" {
		filePath = config.Config.TokenPath
	}

	return ReadFile(filePath)
}

// ReadFile will read the the content of a file and return it as a string.
func ReadFile(filePath string) string {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Unable to read the file at path '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	return fmt.Sprint(string(data))
}
