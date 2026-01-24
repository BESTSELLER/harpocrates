package files

import (
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/adapters/secondary/filesystem"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/rs/zerolog/log"
)

var adapter = filesystem.NewAdapter()

// Read will read the the content of a file and return it as a string.
func Read(filePath string) string {
	data, err := adapter.Read(filePath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to read the file at path '%s'", filePath)
		os.Exit(1)
	}

	return data
}

// Write will write some string data to a file
func Write(output string, fileName string, content interface{}, owner *int, append bool) {
	var ownerToUse *int
	if owner != nil {
		ownerToUse = owner
	} else if config.Config.Owner != -1 {
		ownerToUse = &config.Config.Owner
	}

	err := adapter.Write(output, fileName, content, ownerToUse, append)
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
}
