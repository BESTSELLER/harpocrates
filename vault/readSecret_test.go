package vault

import (
	"fmt"
	"testing"
)

// TestReadSecretWrongPath tests that the function fails on unknown secret path
func TestReadSecretWrongPath(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})
	path := "somepath"

	// act
	vaultClient := &API{
		Client: testClient,
	}
	_, err := vaultClient.ReadSecret(path)
	if err == nil {
		t.Error("expected error got nil")
	}

	// assert
	expected := fmt.Sprintf(secretNotFound, path, nil)
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func testReadSecretKey(path string, key string, expectedValue interface{}, t *testing.T) {
	// mock the ReadSecret function
	vaultClient := &API{
		Client: testClient,
	}

	// act
	value, err := vaultClient.ReadSecretKey(path, key)

	// assert
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != expectedValue {
		t.Errorf("expected value %v, got %v", expectedValue, value)
	}
}

// TestReadSecretKeyWithNumberAsValue tests that the function returns the value as a number
func TestReadSecretKeyWithNumberAsValue(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})
	path := "secret/data/secret"
	key := "key4"
	expectedValue := float64(123)

	testReadSecretKey(path, key, expectedValue, t)
}

// TestReadSecretKeyWithBooleanAsValue tests that the function returns the value as a boolean
func TestReadSecretKeyWithBooleanAsValue(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})
	path := "secret/data/secret"
	key := "key5"
	expectedValue := true

	testReadSecretKey(path, key, expectedValue, t)
}

// TestReadSecretKeyNotFound tests that the function will fail when trying to fetch an unknown key
func TestReadSecretKeyNotFound(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})
	path := "secret/data/secret"
	key := "keys666"

	// mock the ReadSecret function
	vaultClient := &API{
		Client: testClient,
	}

	// act
	_, err := vaultClient.ReadSecretKey(path, key)

	// assert
	if err == nil {
		t.Error("expected error got nil")
	} else {
		expectedMsg := "the key 'keys666' was not found in the path 'secret/data/secret': <nil>"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	}
}
