package vault

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

// TestReadSecretWrongPath tests that the function fails on unknown secret path
func TestReadSecretWrongPath(t *testing.T) {
	// arrange
	path := "somepath"

	// act
	vaultClient := &API{
		Client: testVault.Client,
	}
	_, err := vaultClient.ReadSecret(path)
	if err == nil {
		t.Error("expected error got nil")
	}

	// assert
	// The error message now includes more context due to error wrapping.
	// We'll check if the error string contains the specific part we're interested in.
	expectedErrorSubstring := fmt.Sprintf(secretNotFound, path, "path not found or no content")
	assert.ErrorContains(t, err, expectedErrorSubstring)
}

func testReadSecretKey(path string, key string, expectedValue interface{}, t *testing.T) {
	// mock the ReadSecret function
	vaultClient := &API{
		Client: testVault.Client,
	}

	// act
	value, err := vaultClient.ReadSecretKey(path, key)

	// assert
	assert.NilError(t, err)
	assert.Equal(t, expectedValue, value)
}

// TestReadSecretKeyWithNumberAsValue tests that the function returns the value as a number
func TestReadSecretKeyWithNumberAsValue(t *testing.T) {
	// arrange
	path := "secret/data/secret"
	key := "key4"
	expectedValue := float64(123)

	testReadSecretKey(path, key, expectedValue, t)
}

// TestReadSecretKeyWithBooleanAsValue tests that the function returns the value as a boolean
func TestReadSecretKeyWithBooleanAsValue(t *testing.T) {
	// arrange
	path := "secret/data/secret"
	key := "key5"
	expectedValue := true

	testReadSecretKey(path, key, expectedValue, t)
}

// TestReadSecretKeyNotFound tests that the function will fail when trying to fetch an unknown key
func TestReadSecretKeyNotFound(t *testing.T) {
	// arrange
	path := "secret/data/secret"
	key := "keys666"

	// mock the ReadSecret function
	vaultClient := &API{
		Client: testVault.Client,
	}

	// act
	_, err := vaultClient.ReadSecretKey(path, key)

	// assert
	expectedErrorSubstring := fmt.Sprintf(keyNotFound, key, path, "key does not exist in the secret data")
	assert.Error(t, err, expectedErrorSubstring)
}
