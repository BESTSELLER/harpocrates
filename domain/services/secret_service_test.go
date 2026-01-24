package services_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/domain/ports"
	"github.com/BESTSELLER/harpocrates/domain/services"
	"github.com/BESTSELLER/harpocrates/util"
)

// TestSecretServiceCanBeCreated verifies that the SecretService can be instantiated
func TestSecretServiceCanBeCreated(t *testing.T) {
	// Mock implementations for testing
	mockFetcher := &mockSecretFetcher{}
	mockWriter := &mockSecretWriter{}
	mockAuth := &mockAuthenticator{}
	
	service := services.NewSecretService(mockFetcher, mockWriter, mockAuth)
	
	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}
}

// TestExtractSecretsSimpleSecret tests extraction of a simple secret path
func TestExtractSecretsSimpleSecret(t *testing.T) {
	mockFetcher := &mockSecretFetcher{
		secrets: map[string]map[string]interface{}{
			"secret/data/test": {
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
		},
	}
	mockWriter := &mockSecretWriter{}
	
	service := services.NewSecretService(mockFetcher, mockWriter, nil)
	
	// Setup config
	config.Config.Prefix = "TEST_"
	config.Config.Format = "env"
	
	input := util.SecretJSON{
		Secrets: []interface{}{"secret/data/test"},
		Prefix:  "TEST_",
		Output:  "/tmp/test",
	}
	
	result, err := service.ExtractSecrets(input, false)
	if err != nil {
		t.Fatalf("ExtractSecrets failed: %v", err)
	}
	
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	
	if len(result[0].Result) != 3 {
		t.Errorf("Expected 3 secrets, got %d", len(result[0].Result))
	}
}

// TestExtractSecretsWithKeys tests extraction with specific keys
func TestExtractSecretsWithKeys(t *testing.T) {
	mockFetcher := &mockSecretFetcher{
		secrets: map[string]map[string]interface{}{
			"secret/data/test": {
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		secretKeys: map[string]interface{}{
			"secret/data/test:key1": "value1",
			"secret/data/test:key2": "value2",
		},
	}
	mockWriter := &mockSecretWriter{}
	
	service := services.NewSecretService(mockFetcher, mockWriter, nil)
	
	config.Config.Prefix = "TEST_"
	config.Config.Format = "env"
	
	input := util.SecretJSON{
		Secrets: []interface{}{
			map[string]interface{}{
				"secret/data/test": map[string]interface{}{
					"keys": []interface{}{"key1", "key2"},
				},
			},
		},
		Prefix: "TEST_",
		Output: "/tmp/test",
	}
	
	result, err := service.ExtractSecrets(input, false)
	if err != nil {
		t.Fatalf("ExtractSecrets failed: %v", err)
	}
	
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	
	if len(result[0].Result) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(result[0].Result))
	}
}

// TestExtractSecretsWithPrefix tests custom prefix handling
func TestExtractSecretsWithPrefix(t *testing.T) {
	mockFetcher := &mockSecretFetcher{
		secrets: map[string]map[string]interface{}{
			"secret/data/test": {
				"key1": "value1",
			},
		},
		secretKeys: map[string]interface{}{
			"secret/data/test:key1": "value1",
		},
	}
	mockWriter := &mockSecretWriter{}
	
	service := services.NewSecretService(mockFetcher, mockWriter, nil)
	
	config.Config.Prefix = "DEFAULT_"
	config.Config.Format = "env"
	
	input := util.SecretJSON{
		Secrets: []interface{}{
			map[string]interface{}{
				"secret/data/test": map[string]interface{}{
					"prefix": "CUSTOM_",
					"keys":   []interface{}{"key1"},
				},
			},
		},
		Output: "/tmp/test",
	}
	
	result, err := service.ExtractSecrets(input, false)
	if err != nil {
		t.Fatalf("ExtractSecrets failed: %v", err)
	}
	
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	
	// Check that custom prefix was used
	hasCustomPrefix := false
	for k := range result[0].Result {
		if len(k) >= 7 && k[:7] == "CUSTOM_" {
			hasCustomPrefix = true
			break
		}
	}
	
	if !hasCustomPrefix {
		t.Errorf("Expected custom prefix CUSTOM_ to be used")
	}
}

// TestExtractSecretsErrorHandling tests error propagation
func TestExtractSecretsErrorHandling(t *testing.T) {
	mockFetcher := &mockSecretFetcher{
		shouldError: true,
	}
	mockWriter := &mockSecretWriter{}
	
	service := services.NewSecretService(mockFetcher, mockWriter, nil)
	
	config.Config.Prefix = "TEST_"
	
	input := util.SecretJSON{
		Secrets: []interface{}{"secret/data/nonexistent"},
		Output:  "/tmp/test",
	}
	
	_, err := service.ExtractSecrets(input, false)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// Mock implementations for testing
type mockSecretFetcher struct {
	secrets     map[string]map[string]interface{}
	secretKeys  map[string]interface{}
	shouldError bool
}

func (m *mockSecretFetcher) ReadSecret(path string) (map[string]interface{}, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	if m.secrets != nil {
		if secret, ok := m.secrets[path]; ok {
			return secret, nil
		}
	}
	return map[string]interface{}{"key": "value"}, nil
}

func (m *mockSecretFetcher) ReadSecretKey(path string, key string) (interface{}, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	if m.secretKeys != nil {
		lookupKey := fmt.Sprintf("%s:%s", path, key)
		if value, ok := m.secretKeys[lookupKey]; ok {
			return value, nil
		}
	}
	return "value", nil
}

type mockSecretWriter struct {
	writeCalls []writeCall
}

type writeCall struct {
	output   string
	fileName string
	content  interface{}
}

func (m *mockSecretWriter) Write(output string, fileName string, content interface{}, owner *int, appendToFile bool) error {
	m.writeCalls = append(m.writeCalls, writeCall{
		output:   output,
		fileName: fileName,
		content:  content,
	})
	return nil
}

func (m *mockSecretWriter) Read(filePath string) (string, error) {
	return "content", nil
}

type mockAuthenticator struct{}

func (m *mockAuthenticator) Login() (*ports.AuthResult, error) {
	return nil, nil
}

func (m *mockAuthenticator) IsTokenValid(token string, expiry time.Time) bool {
	return true
}
