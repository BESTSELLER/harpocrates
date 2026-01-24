package vault_test

import (
	"context"
	"testing"

	vaultadapter "github.com/BESTSELLER/harpocrates/adapters/secondary/vault"
	api "github.com/hashicorp/vault/api"
	"github.com/testcontainers/testcontainers-go/modules/vault"
)

// TestVaultAdapterReadSecret tests the ReadSecret functionality with a real Vault instance
func TestVaultAdapterReadSecret(t *testing.T) {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:latest",
		vault.WithToken("testtoken"),
	)
	if err != nil {
		t.Fatalf("failed to start vault container: %s", err)
	}
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate vault container: %s", err)
		}
	}()

	// Get Vault connection details
	vaultAddr, err := vaultContainer.HttpHostAddress(ctx)
	if err != nil {
		t.Fatalf("failed to get vault address: %s", err)
	}

	// Create Vault client
	client, err := api.NewClient(&api.Config{
		Address: vaultAddr,
	})
	if err != nil {
		t.Fatalf("failed to create vault client: %s", err)
	}
	client.SetToken("testtoken")

	// Write test data
	testData := map[string]interface{}{
		"data": map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		},
	}
	_, err = client.Logical().Write("secret/data/test", testData)
	if err != nil {
		t.Fatalf("failed to write test data: %s", err)
	}

	// Create adapter
	adapter := vaultadapter.NewAdapter(client)

	// Test ReadSecret
	result, err := adapter.ReadSecret("secret/data/test")
	if err != nil {
		t.Fatalf("ReadSecret failed: %s", err)
	}

	if result["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", result["key1"])
	}
	if result["key2"].(float64) != 42 {
		t.Errorf("Expected key2=42, got %v", result["key2"])
	}
	if result["key3"] != true {
		t.Errorf("Expected key3=true, got %v", result["key3"])
	}
}

// TestVaultAdapterReadSecretKeyNested tests nested key traversal
func TestVaultAdapterReadSecretKeyNested(t *testing.T) {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:latest",
		vault.WithToken("testtoken"),
	)
	if err != nil {
		t.Fatalf("failed to start vault container: %s", err)
	}
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate vault container: %s", err)
		}
	}()

	vaultAddr, err := vaultContainer.HttpHostAddress(ctx)
	if err != nil {
		t.Fatalf("failed to get vault address: %s", err)
	}

	client, err := api.NewClient(&api.Config{
		Address: vaultAddr,
	})
	if err != nil {
		t.Fatalf("failed to create vault client: %s", err)
	}
	client.SetToken("testtoken")

	// Write nested test data
	testData := map[string]interface{}{
		"data": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "nested_value",
			},
			"users": []interface{}{
				map[string]interface{}{"name": "Alice"},
				map[string]interface{}{"name": "Bob"},
			},
		},
	}
	_, err = client.Logical().Write("secret/data/nested", testData)
	if err != nil {
		t.Fatalf("failed to write test data: %s", err)
	}

	adapter := vaultadapter.NewAdapter(client)

	// Test nested key access
	result, err := adapter.ReadSecretKey("secret/data/nested", "level1.level2")
	if err != nil {
		t.Fatalf("ReadSecretKey failed: %s", err)
	}
	if result != "nested_value" {
		t.Errorf("Expected nested_value, got %v", result)
	}

	// Test array access
	result, err = adapter.ReadSecretKey("secret/data/nested", "users[0].name")
	if err != nil {
		t.Fatalf("ReadSecretKey array access failed: %s", err)
	}
	if result != "Alice" {
		t.Errorf("Expected Alice, got %v", result)
	}

	result, err = adapter.ReadSecretKey("secret/data/nested", "users[1].name")
	if err != nil {
		t.Fatalf("ReadSecretKey array access failed: %s", err)
	}
	if result != "Bob" {
		t.Errorf("Expected Bob, got %v", result)
	}
}

// TestVaultAdapterReadSecretKeyNotFound tests error handling
func TestVaultAdapterReadSecretKeyNotFound(t *testing.T) {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:latest",
		vault.WithToken("testtoken"),
	)
	if err != nil {
		t.Fatalf("failed to start vault container: %s", err)
	}
	defer func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate vault container: %s", err)
		}
	}()

	vaultAddr, err := vaultContainer.HttpHostAddress(ctx)
	if err != nil {
		t.Fatalf("failed to get vault address: %s", err)
	}

	client, err := api.NewClient(&api.Config{
		Address: vaultAddr,
	})
	if err != nil {
		t.Fatalf("failed to create vault client: %s", err)
	}
	client.SetToken("testtoken")

	testData := map[string]interface{}{
		"data": map[string]interface{}{
			"key1": "value1",
		},
	}
	_, err = client.Logical().Write("secret/data/test", testData)
	if err != nil {
		t.Fatalf("failed to write test data: %s", err)
	}

	adapter := vaultadapter.NewAdapter(client)

	// Test non-existent key
	_, err = adapter.ReadSecretKey("secret/data/test", "nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent key, got nil")
	}
}
