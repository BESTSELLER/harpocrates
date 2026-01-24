package files

import (
	"os"
	"sync"

	"github.com/BESTSELLER/harpocrates/adapters/secondary/filesystem"
	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/domain/ports"
	"github.com/rs/zerolog/log"
)

var (
	adapter     ports.SecretWriter
	adapterOnce sync.Once
	adapterMu   sync.RWMutex
)

// getAdapter returns the adapter, initializing it if necessary
func getAdapter() ports.SecretWriter {
	adapterMu.RLock()
	if adapter != nil {
		defer adapterMu.RUnlock()
		return adapter
	}
	adapterMu.RUnlock()
	
	adapterOnce.Do(func() {
		adapterMu.Lock()
		adapter = filesystem.NewAdapter()
		adapterMu.Unlock()
	})
	
	adapterMu.RLock()
	defer adapterMu.RUnlock()
	return adapter
}

// SetAdapter allows injecting a custom adapter for testing purposes
func SetAdapter(a ports.SecretWriter) {
	adapterMu.Lock()
	defer adapterMu.Unlock()
	adapter = a
}

// ResetAdapter resets the adapter to nil (useful for testing)
func ResetAdapter() {
	adapterMu.Lock()
	defer adapterMu.Unlock()
	adapter = nil
	adapterOnce = sync.Once{}
}

// Read will read the the content of a file and return it as a string.
func Read(filePath string) string {
	data, err := getAdapter().Read(filePath)
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

	err := getAdapter().Write(output, fileName, content, ownerToUse, append)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write file")
		os.Exit(1)
	}
}
