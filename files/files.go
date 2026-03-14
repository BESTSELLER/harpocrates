package files

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/rs/zerolog/log"
)

// Read will read the content of a file and return it as a string.
func Read(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to read the file at path '%s'", filePath)
	}

	return fmt.Sprint(string(data))
}

// Write will write some string data to a file
func Write(output string, fileName string, content interface{}, owner *int, append bool) {
	fileName = fixFileName(fileName)
	path := filepath.Join(output, fileName)

	if _, err := os.Stat(output); os.IsNotExist(err) {
		err = os.MkdirAll(output, 0700)
		if err != nil {
			log.Fatal().Err(err).Msgf("Unable to create dir at path '%s'", output)
		}
	}

	overWriteOrAppend := os.O_TRUNC
	if append {
		overWriteOrAppend = os.O_APPEND
	}

	f, err := os.OpenFile(path, overWriteOrAppend|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal().Err(err).Msgf("An error happened while trying to open file %s", path)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal().Err(err).Msgf("Unable to close file '%s'", path)
		}
	}()

	if _, err = fmt.Fprintf(f, "%v", content); err != nil {
		log.Fatal().Err(err).Msgf("Unable to write to file '%s'", path)
	}
	log.Debug().Msgf("Wrote file '%s'", path)

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
		log.Fatal().Err(err).Msgf("Unable to set permissions to folder '%s'", path)
	}

	if err := f.Chown(owner, -1); err != nil {
		log.Fatal().Err(err).Msgf("Unable to set permissions to file '%s'", path)
	}
}

func fixFileName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9.-]+")
	fileName := reg.ReplaceAllString(name, "_")

	return fileName
}
