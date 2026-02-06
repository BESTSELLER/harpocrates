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

// TestReadSecretKeyNested tests nested key fetching scenarios
func TestReadSecretKeyNested(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	path := "secret/data/complex"

	// Define the complex secret structure matching your requirements
	secretData := map[string]interface{}{
		"globalSecrets": map[string]interface{}{
			"theSecretINeed": "HelloThere!",
		},
		"list": []interface{}{"item1", "item2"},
		"users": []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
		},
		"key[with]brackets": "literalValue",
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "deepValue",
			},
		},
	}

	// Write the secret to Vault
	// We wrap secretData in "data" because the test setup uses KV v2 secret engine
	_, err := testClient.Logical().Write(path, map[string]interface{}{
		"data": secretData,
	})
	if err != nil {
		t.Fatalf("failed to write secret data to vault: %v", err)
	}

	// Act & Assert

	// 1. Nested map access: globalSecrets.theSecretINeed
	testReadSecretKey(path, "globalSecrets.theSecretINeed", "HelloThere!", t)

	// 2. Array access: list[0]
	testReadSecretKey(path, "list[0]", "item1", t)

	// 3. Nested array access: users[0].name
	testReadSecretKey(path, "users[0].name", "Alice", t)
	testReadSecretKey(path, "users.0.name", "Alice", t)

	// 4. Literal key with brackets (regression/feature): "key[with]brackets"
	testReadSecretKey(path, "key[with]brackets", "literalValue", t)

	// 5. Deep nesting: a.b.c
	testReadSecretKey(path, "a.b.c", "deepValue", t)
}

// TestReadSecretKeyDeeplyNested tests retrieval of a key nested 7 levels deep with mixed maps and lists
func TestReadSecretKeyDeeplyNested(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	path := "secret/data/deep"

	// Construct 7 levels of nesting (mixed maps and lists)
	// Path: level1[0].level3.level4[1].level6.final
	secretData := map[string]interface{}{
		"level1": []interface{}{
			map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": []interface{}{
						"dummy", // index 0
						map[string]interface{}{ // index 1
							"level6": map[string]interface{}{
								"final": "FoundIt",
							},
						},
					},
				},
			},
		},
	}

	// Write the secret to Vault
	_, err := testClient.Logical().Write(path, map[string]interface{}{
		"data": secretData,
	})
	if err != nil {
		t.Fatalf("failed to write secret data to vault: %v", err)
	}

	// act & assert
	testReadSecretKey(path, "level1[0].level3.level4[1].level6.final", "FoundIt", t)
}

// TestReadSecretKeyLiteralWithDots confirms that keys with literal dots are found before dot-notation traversal
func TestReadSecretKeyLiteralWithDots(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	path := "secret/data/literal_dots"
	secretData := map[string]interface{}{
		"key.with.dots": "literalValue",
		"key": map[string]interface{}{
			"with": map[string]interface{}{
				"dots":         "nestedValue",
				"dots1.nested": "nextNestedValues",
				"this.should.also": map[string]interface{}{
					"work": "right?",
				},
			},
		},
	}

	_, err := testClient.Logical().Write(path, map[string]interface{}{
		"data": secretData,
	})
	if err != nil {
		t.Fatalf("failed to write secret data to vault: %v", err)
	}

	// act & assert
	// Should prioritize literal match ("literalValue") over nested traversal ("nestedValue")
	testReadSecretKey(path, "key.with.dots", "literalValue", t)
	testReadSecretKey(path, "key.with.dots1.nested", "nextNestedValues", t)
	testReadSecretKey(path, "key.with.this.should.also.work", "right?", t)
}

// TestReadSecretKeyNestedNotFound tests that accessing a non-existent nested key returns an error
func TestReadSecretKeyNestedNotFound(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	path := "secret/data/nested_not_found"
	secretData := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": "value",
		},
	}

	_, err := testClient.Logical().Write(path, map[string]interface{}{
		"data": secretData,
	})
	if err != nil {
		t.Fatalf("failed to write secret data to vault: %v", err)
	}

	// act
	// accessing "parent.missing" -> should fail because "missing" is not in {child: value}
	vaultClient := &API{Client: testClient}
	key := "parent.missing"
	val, err := vaultClient.ReadSecretKey(path, key)

	// assert
	if err == nil {
		t.Errorf("expected error for key %q, got value: %v", key, val)
	}
}
