package util

import (
	"testing"
)

// TestIsSecretString tests the IsSecretString helper function
func TestIsSecretString(t *testing.T) {
	tests := []struct {
		name     string
		item     SecretItem
		expected bool
	}{
		{
			name:     "string type",
			item:     "secret/data/test",
			expected: true,
		},
		{
			name:     "map type",
			item:     map[string]any{"key": "value"},
			expected: false,
		},
		{
			name:     "nil",
			item:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSecretString(tt.item)
			if result != tt.expected {
				t.Errorf("IsSecretString(%v) = %v, want %v", tt.item, result, tt.expected)
			}
		})
	}
}

// TestIsKeyString tests the IsKeyString helper function
func TestIsKeyString(t *testing.T) {
	tests := []struct {
		name     string
		item     KeyItem
		expected bool
	}{
		{
			name:     "string type",
			item:     "key1",
			expected: true,
		},
		{
			name:     "map type",
			item:     map[string]any{"prefix": "TEST_"},
			expected: false,
		},
		{
			name:     "nil",
			item:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKeyString(tt.item)
			if result != tt.expected {
				t.Errorf("IsKeyString(%v) = %v, want %v", tt.item, result, tt.expected)
			}
		})
	}
}

// TestGetSecretString tests the GetSecretString helper function
func TestGetSecretString(t *testing.T) {
	tests := []struct {
		name           string
		item           SecretItem
		expectedString string
		expectedOk     bool
	}{
		{
			name:           "valid string",
			item:           "secret/data/test",
			expectedString: "secret/data/test",
			expectedOk:     true,
		},
		{
			name:           "map type",
			item:           map[string]any{"key": "value"},
			expectedString: "",
			expectedOk:     false,
		},
		{
			name:           "nil",
			item:           nil,
			expectedString: "",
			expectedOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, ok := GetSecretString(tt.item)
			if str != tt.expectedString || ok != tt.expectedOk {
				t.Errorf("GetSecretString(%v) = (%v, %v), want (%v, %v)", 
					tt.item, str, ok, tt.expectedString, tt.expectedOk)
			}
		})
	}
}

// TestGetKeyString tests the GetKeyString helper function
func TestGetKeyString(t *testing.T) {
	tests := []struct {
		name           string
		item           KeyItem
		expectedString string
		expectedOk     bool
	}{
		{
			name:           "valid string",
			item:           "key1",
			expectedString: "key1",
			expectedOk:     true,
		},
		{
			name:           "map type",
			item:           map[string]any{"prefix": "TEST_"},
			expectedString: "",
			expectedOk:     false,
		},
		{
			name:           "nil",
			item:           nil,
			expectedString: "",
			expectedOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, ok := GetKeyString(tt.item)
			if str != tt.expectedString || ok != tt.expectedOk {
				t.Errorf("GetKeyString(%v) = (%v, %v), want (%v, %v)", 
					tt.item, str, ok, tt.expectedString, tt.expectedOk)
			}
		})
	}
}

// TestGetSecretMap tests the GetSecretMap helper function
func TestGetSecretMap(t *testing.T) {
	tests := []struct {
		name       string
		item       SecretItem
		expectedOk bool
	}{
		{
			name:       "valid map",
			item:       map[string]any{"format": "json", "filename": "test.json"},
			expectedOk: true,
		},
		{
			name:       "string type",
			item:       "secret/data/test",
			expectedOk: false,
		},
		{
			name:       "nil",
			item:       nil,
			expectedOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ok := GetSecretMap(tt.item)
			if ok != tt.expectedOk {
				t.Errorf("GetSecretMap(%v) ok = %v, want %v", tt.item, ok, tt.expectedOk)
			}
			if ok && m == nil {
				t.Errorf("GetSecretMap(%v) returned nil map when ok=true", tt.item)
			}
			if !ok && m != nil {
				t.Errorf("GetSecretMap(%v) returned non-nil map when ok=false", tt.item)
			}
		})
	}
}

// TestGetKeyMap tests the GetKeyMap helper function
func TestGetKeyMap(t *testing.T) {
	tests := []struct {
		name       string
		item       KeyItem
		expectedOk bool
	}{
		{
			name:       "valid map",
			item:       map[string]any{"prefix": "TEST_", "uppercase": true},
			expectedOk: true,
		},
		{
			name:       "string type",
			item:       "key1",
			expectedOk: false,
		},
		{
			name:       "nil",
			item:       nil,
			expectedOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ok := GetKeyMap(tt.item)
			if ok != tt.expectedOk {
				t.Errorf("GetKeyMap(%v) ok = %v, want %v", tt.item, ok, tt.expectedOk)
			}
			if ok && m == nil {
				t.Errorf("GetKeyMap(%v) returned nil map when ok=true", tt.item)
			}
			if !ok && m != nil {
				t.Errorf("GetKeyMap(%v) returned non-nil map when ok=false", tt.item)
			}
		})
	}
}

// TestSecretItemUsage tests the usage of SecretItem in SecretJSON
func TestSecretItemUsage(t *testing.T) {
	// Test that SecretJSON can hold both string and map secrets
	secretJSON := SecretJSON{
		Secrets: []SecretItem{
			"secret/data/simple",
			map[string]any{
				"secret/data/complex": map[string]any{
					"format":   "json",
					"filename": "test.json",
				},
			},
		},
	}

	if len(secretJSON.Secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(secretJSON.Secrets))
	}

	// Verify first item is a string
	if !IsSecretString(secretJSON.Secrets[0]) {
		t.Error("First secret should be a string")
	}

	// Verify second item is a map
	if IsSecretString(secretJSON.Secrets[1]) {
		t.Error("Second secret should not be a string")
	}
}

// TestKeyItemUsage tests the usage of KeyItem in Secret
func TestKeyItemUsage(t *testing.T) {
	// Test that Secret can hold both string and map keys
	secret := Secret{
		Keys: []KeyItem{
			"key1",
			map[string]any{
				"key2": map[string]any{
					"prefix": "TEST_",
				},
			},
		},
	}

	if len(secret.Keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(secret.Keys))
	}

	// Verify first item is a string
	if !IsKeyString(secret.Keys[0]) {
		t.Error("First key should be a string")
	}

	// Verify second item is a map
	if IsKeyString(secret.Keys[1]) {
		t.Error("Second key should not be a string")
	}
}
