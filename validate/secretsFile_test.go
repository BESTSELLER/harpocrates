package validate

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
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
		// If it's a folder or file isn't a yaml then just skip it.
		if test.IsDir() || !strings.Contains(test.Name(), ".yaml") || !strings.Contains(test.Name(), ".yml") {
			continue
		}
		collectedTests = append(collectedTests, fmt.Sprintf("../test_data/%s", test.Name()))
	}

	return collectedTests
}

// Test all secret file tests in the test_data directory.
func TestSecretsFile(t *testing.T) {
	tests := collectTests()
	
	for _, test := range tests {
		switch strings.Contains(test, "not_valid") {
		case true:
			if SecretsFile(test, "./schema.json") {
				t.Errorf("Expected %s to fail validation", test)
			}
		case false:
			if !SecretsFile(test, "./schema.json") {
				t.Errorf("Expected %s to pass validation", test)
			}
		}
	}
}
