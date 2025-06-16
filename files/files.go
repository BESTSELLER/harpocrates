package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/rs/zerolog/log"
)

// Read will read the content of a file and return it as a string.
func Read(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read file at path '%s': %w", filePath, err)
	}
	return string(data), nil // Removed fmt.Sprint, as string(data) is already a string
}

// Write will write string data to a file.
func Write(outputDir string, fileName string, content interface{}, owner *int, appendToFile bool) error {
	fileName = fixFileName(fileName)
	filePath := filepath.Join(outputDir, fileName)

	// Ensure the output directory exists.
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0700); err != nil {
			return fmt.Errorf("unable to create dir at path '%s': %w", outputDir, err)
		}
	}

	// Determine file opening mode (truncate or append).
	fileFlags := os.O_CREATE | os.O_WRONLY
	if appendToFile {
		fileFlags |= os.O_APPEND
	} else {
		fileFlags |= os.O_TRUNC
	}

	f, err := os.OpenFile(filePath, fileFlags, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer f.Close()

	// Write content to the file.
	if _, err = f.WriteString(fmt.Sprintf("%v", content)); err != nil {
		return fmt.Errorf("unable to write to file '%s': %w", filePath, err)
	}
	log.Debug().Msgf("Wrote file '%s'", filePath)

	// Set permissions on the file and folder.
	effectiveOwner := -1 // Default to no change if not specified.
	if owner != nil {
		effectiveOwner = *owner
	} else if config.Config.Owner != -1 {
		effectiveOwner = config.Config.Owner
	}

	if effectiveOwner != -1 {
		if err := setPermissions(f, filePath, outputDir, effectiveOwner); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}
	return nil
}

// setPermissions applies ownership to the file and its output directory.
func setPermissions(f *os.File, filePath string, outputDir string, ownerUID int) error {
	// Set ownership for the output directory.
	// Note: Chown changes the numeric uid and gid of the named file.
	// For directories, it's common to only change UID if GID isn't specified (-1 means no change).
	if err := os.Chown(outputDir, ownerUID, -1); err != nil {
		return fmt.Errorf("unable to set permissions for folder '%s': %w", outputDir, err)
	}

	// Set ownership for the file.
	if err := f.Chown(ownerUID, -1); err != nil {
		return fmt.Errorf("unable to set permissions for file '%s': %w", filePath, err)
	}
	return nil
}

func fixFileName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9.-]+")
	fileName := reg.ReplaceAllString(name, "_")

	return fileName
}
