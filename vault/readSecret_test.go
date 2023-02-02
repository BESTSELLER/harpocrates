package vault

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
)

// TestReadSecretWrongPath tests that the function fails on unknown secret path
func TestReadSecretWrongPath(t *testing.T) {
	// arrange
	path := "somepath"

	// act
	_, err := NewClient().ReadSecret(path)
	if err == nil {
		t.Error("expected error got nil")
	}

	// assert
	assert.Equal(t, fmt.Sprintf(secretNotFound, path, err), err.Error())
}
