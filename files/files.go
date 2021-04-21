package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BESTSELLER/harpocrates/config"
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
func Write(output string, fileName string, content string, owner *int) {
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
		os.Exit(1)
	}

	// set permissions on file and folder
	if owner != nil {
		setPermissions(f, path, output, *owner)
		return
	}
	if config.Config.Owner != -1 {
		setPermissions(f, path, output, config.Config.Owner)
		return
	}
}

func setPermissions(f *os.File, path string, output string, owner int) {
	if err := os.Chown(output, owner, -1); err != nil {
		fmt.Printf("Unable to set permissions to folder '%s': %v\n", path, err)
		os.Exit(1)
	}

	if err := f.Chown(owner, -1); err != nil {
		fmt.Printf("Unable to set permissions to file '%s': %v\n", path, err)
		os.Exit(1)
	}
}

func fixFileName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9.-]+")
	fileName := reg.ReplaceAllString(name, "_")

	return fileName
}
