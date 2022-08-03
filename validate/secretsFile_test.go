package validate

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/BESTSELLER/harpocrates/files"
	"github.com/rs/zerolog/log"
)

// Collect all secret file tests from the test_data directory.
// Return as a slice of strings.
func collectTests() []string {
	tests, err := ioutil.ReadDir("../test_data")
	if err != nil {
		panic(err)
	}

	collectedTests := make([]string, 0)
	for _, test := range tests {
		// Only include if it's yaml or json
		if strings.Contains(test.Name(), ".yaml") || strings.Contains(test.Name(), ".yml") || strings.Contains(test.Name(), ".json") {
			collectedTests = append(collectedTests, fmt.Sprintf("../test_data/%s", test.Name()))
			continue
		}
	}
	return collectedTests
}

// Test all secret file tests in the test_data directory.
func TestSecretsFile(t *testing.T) {
	tests := collectTests()
	for _, test := range tests {
		log.Debug().Msgf("Testing file: %s\n", test)
		file := files.Read(test)
		switch strings.Contains(test, "not_valid") {
		case true:
			if SecretsFile(file) {
				t.Errorf("Expected %s to fail validation", test)
			}
		case false:
			if !SecretsFile(file) {
				t.Errorf("Expected %s to pass validation", test)
			}
		}
	}
}
