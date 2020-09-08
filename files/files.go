package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

// Read will read the the content of a file and return it as a string.
func Read(filePath string) string {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Unable to read the file at path '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	return fmt.Sprint(string(data))
}

// Write will write some string data to a file
func Write(output string, fileName string, content string) {
	fileName = fixFileName(fileName)
	path := filepath.Join(output, fileName)

	if _, err := os.Stat(output); os.IsNotExist(err) {
		err = os.MkdirAll(output, 0700)
		if err != nil {
			fmt.Printf("Unable to create dir at path '%s': %v\n", output, err)
			os.Exit(1)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("An error happened while trying to open file %s: %s\n", path, err)
		os.Exit(1)
	}

	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		fmt.Printf("Unable to write to file '%s': %v\n", path, err)
		f.Close()
		os.Exit(1)
	}
}

func fixFileName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9.]+")
	fileName := reg.ReplaceAllString(name, "_")

	return fileName
}
