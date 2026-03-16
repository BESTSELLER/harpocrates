package vault

import (
	"os"
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

func TestExtractSecretsWithOverrideNameAndSaveAsFile(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	// define input
	data := files.Read("../test_data/override_name_save_as_file.yaml")
	input := util.ReadInput(data)

	// mock prefix
	config.Config.Prefix = input.Prefix

	vaultClient := &API{
		Client: testClient,
	}

	// clean up: file should be written using the overridden name, not the original key
	defer os.Remove("../.tmp/TEST_newKey1") //nolint:errcheck // It's just tests, we don't care
	defer os.Remove("../.tmp/TEST_key1")    //nolint:errcheck // It's just tests, we don't care

	// act
	_, err := vaultClient.ExtractSecrets(input, false)
	if err != nil {
		t.Error(err)
	}

	// assert: file must exist with the overridden name
	content, err := os.ReadFile("../.tmp/TEST_newKey1")
	if err != nil {
		t.Errorf("expected file ../.tmp/TEST_newKey1 to exist (overridden name), but got error: %v", err)
	}

	expected := "value1"
	actual := string(content)
	if expected != actual {
		t.Errorf("expected file content %q, got %q", expected, actual)
	}

	// assert: file must NOT exist with the original key name
	if _, err := os.Stat("../.tmp/TEST_key1"); err == nil {
		t.Errorf("expected no file ../.tmp/TEST_key1 (original key), but it was found")
	}
}

func TestExtractSecretsWithKeyLevelUpperCase(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	// define input
	data := files.Read("../test_data/override_name_uppercase.yaml")
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

	// assert: key1 has uppercase: true at the key level, so it should be uppercased to KEY1
	for _, v := range result {
		resMap := v.Result
		if val, ok := resMap["KEY1"]; !ok {
			t.Errorf("Expected 'KEY1' (uppercased) in result, got %v", resMap)
		} else if val != "value1" {
			t.Errorf("Expected 'KEY1' to be 'value1', got '%v'", val)
		}

		// key1 (lowercase) should NOT be present
		if _, ok := resMap["key1"]; ok {
			t.Errorf("Did not expect 'key1' (lowercase) in result when uppercase is true, got %v", resMap)
		}

		// key2 has no uppercase config, so it should remain lowercase
		if val, ok := resMap["key2"]; !ok {
			t.Errorf("Expected 'key2' in result, got %v", resMap)
		} else if val != "value2" {
			t.Errorf("Expected 'key2' to be 'value2', got '%v'", val)
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
