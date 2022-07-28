package validate

import (
	"fmt"
	"io/ioutil"
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
		if test.IsDir() {
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
		if !SecretsFile(test, "./schema.json") {
			t.Errorf("SecretsFile(%q) = false", test)
		}
	}
}
