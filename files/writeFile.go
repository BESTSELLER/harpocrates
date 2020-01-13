package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

// WriteFile will write some string data to a file
func WriteFile(dirPath string, fileName string, content string) {
	fileName = fixFileName(fileName)
	path := filepath.Join(dirPath, fileName)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0700)
		if err != nil {
			fmt.Printf("Unable to create dir at path '%s': %v\n", dirPath, err)
			os.Exit(1)
		}
	}

	fileContent := []byte(content)
	err := ioutil.WriteFile(path, fileContent, 0644)
	if err != nil {
		fmt.Printf("Unable to write to file '%s': %v\n", path, err)
		os.Exit(1)
	}
}

func fixFileName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9.]+")
	fileName := reg.ReplaceAllString(name, "_")

	return fileName
}
