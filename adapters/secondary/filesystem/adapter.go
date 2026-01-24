package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BESTSELLER/harpocrates/domain/ports"
	"github.com/rs/zerolog/log"
)

// Adapter implements the SecretWriter port for filesystem operations
type Adapter struct{}

// NewAdapter creates a new filesystem adapter
func NewAdapter() ports.SecretWriter {
	return &Adapter{}
}

// Read reads the content of a file
func (a *Adapter) Read(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read the file at path '%s': %w", filePath, err)
	}

	return string(data), nil
}

// Write writes content to a file
func (a *Adapter) Write(output string, fileName string, content interface{}, owner *int, append bool) error {
	fileName = a.fixFileName(fileName)
	path := filepath.Join(output, fileName)

	if _, err := os.Stat(output); os.IsNotExist(err) {
		err = os.MkdirAll(output, 0700)
		if err != nil {
			return fmt.Errorf("unable to create dir at path '%s': %w", output, err)
		}
	}

	overWriteOrAppend := os.O_TRUNC
	if append {
		overWriteOrAppend = os.O_APPEND
	}

	f, err := os.OpenFile(path, overWriteOrAppend|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("an error happened while trying to open file %s: %w", path, err)
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("%v", content)); err != nil {
		return fmt.Errorf("unable to write to file '%s': %w", path, err)
	}
	log.Debug().Msgf("Wrote file '%s'", path)

	// set permissions on file and folder
	if owner != nil {
		return a.setPermissions(f, path, output, *owner)
	}

	return nil
}

func (a *Adapter) setPermissions(f *os.File, path string, output string, owner int) error {
	if err := os.Chown(output, owner, -1); err != nil {
		return fmt.Errorf("unable to set permissions to folder '%s': %w", path, err)
	}

	if err := f.Chown(owner, -1); err != nil {
		return fmt.Errorf("unable to set permissions to file '%s': %w", path, err)
	}

	return nil
}

func (a *Adapter) fixFileName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9.-]+")
	fileName := reg.ReplaceAllString(name, "_")

	return fileName
}
