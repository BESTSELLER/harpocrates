package files

import (
	"fmt"
	"io/ioutil"
	"os"
)

// ReadTokenFile will read the token from Kubernetes service account
func ReadTokenFile() string {
	filePath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Unable to read the file at path '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	return fmt.Sprint(string(data))
}
