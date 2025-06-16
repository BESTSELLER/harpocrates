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
	for _, testPath := range tests {
		t.Run(testPath, func(t *testing.T) { // Use t.Run for better test output
			log.Debug().Msgf("Testing file: %s", testPath)
			fileContent, err := files.Read(testPath)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", testPath, err)
			}

			validationErr := SecretsFile(fileContent)

			isNotValidTest := strings.Contains(testPath, "not_valid")

			if isNotValidTest {
				if validationErr == nil {
					t.Errorf("Expected %s to fail validation, but it passed", testPath)
				} else {
					// Optionally, check if the error is of the expected type
					if _, ok := validationErr.(*ErrValidationFailed); !ok && !strings.Contains(validationErr.Error(), "failed to convert YAML to JSON") {
						// Allow other errors like YAMLToJSON conversion for files that are not even valid YAML/JSON
						t.Logf("Test %s failed as expected. Error type: %T, Error: %v", testPath, validationErr, validationErr)
					}
				}
			} else { // Expected to be valid
				if validationErr != nil {
					t.Errorf("Expected %s to pass validation, but it failed: %v", testPath, validationErr)
				}
			}
		})
	}
}
