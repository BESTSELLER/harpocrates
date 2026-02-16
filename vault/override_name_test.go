package vault

import (
	"testing"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
)

func TestExtractSecretsWithOverrideName(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	// define input
	data := files.Read("../test_data/override_name.yaml")
	input := util.ReadInput(data)

	// mock prefix
	config.Config.Prefix = input.Prefix

	vaultClient := &API{
		Client: testClient,
	}

	// act
	result, err := vaultClient.ExtractSecrets(input, false)
	if err != nil {
		t.Error(err)
	}

	if len(result) == 0 {
		t.Fatal("No secrets extracted")
	}

	// assert
	for _, v := range result {
		// Assert newKey1 is present and key1 is not (or rather key1 is mapped to value1 but key in result is newKey1)
		resMap := v.Result
		if val, ok := resMap["newKey1"]; !ok {
			t.Errorf("Expected 'newKey1' in result, got %v", resMap)
		} else if val != "value1" {
			t.Errorf("Expected 'newKey1' to be 'value1', got '%v'", val)
		}

		if val, ok := resMap["key2"]; !ok {
			t.Errorf("Expected 'key2' in result, got %v", resMap)
		} else if val != "value2" {
			t.Errorf("Expected 'key2' to be 'value2', got '%v'", val)
		}

		// key1 should not be a key in result
		if _, ok := resMap["key1"]; ok {
			t.Errorf("Did not expect 'key1' in result, got %v", resMap)
		}
	}
}

func TestExtractSecretsWithOverrideNameNested(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	path := "secret/data/complex"
	secretData := map[string]interface{}{
		"globalSecrets": map[string]interface{}{
			"theSecretINeed": "HelloThere!",
		},
	}

	_, err := testClient.Logical().Write(path, map[string]interface{}{
		"data": secretData,
	})
	if err != nil {
		t.Fatalf("failed to write secret data to vault: %v", err)
	}

	// define input
	data := files.Read("../test_data/override_name_nested.yaml")
	input := util.ReadInput(data)

	// mock prefix
	config.Config.Prefix = input.Prefix

	vaultClient := &API{
		Client: testClient,
	}

	// act
	result, err := vaultClient.ExtractSecrets(input, false)
	if err != nil {
		t.Error(err)
	}

	if len(result) == 0 {
		t.Fatal("No secrets extracted")
	}

	// assert
	for _, v := range result {
		resMap := v.Result
		if val, ok := resMap["myTopSecret"]; !ok {
			t.Errorf("Expected 'myTopSecret' in result, got %v", resMap)
		} else if val != "HelloThere!" {
			t.Errorf("Expected 'myTopSecret' to be 'HelloThere!', got '%v'", val)
		}

		if _, ok := resMap["globalSecrets.theSecretINeed"]; ok {
			t.Errorf("Did not expect 'globalSecrets.theSecretINeed' in result, got %v", resMap)
		}
	}
}
