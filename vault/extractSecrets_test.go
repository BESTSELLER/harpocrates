package vault

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/hashicorp/vault/api"
	"github.com/testcontainers/testcontainers-go/modules/vault"
	"gotest.tools/v3/assert"
)

var testClient *api.Client

func setupVault(t *testing.T) {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:latest",
		vault.WithToken("unittesttoken"),
		// vault.WithInitCommand("secrets enable -path=secret kv-v2"),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	t.Cleanup(func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	httpHostAddress, err := vaultContainer.HttpHostAddress(ctx)
	if err != nil {
		t.Fatalf("failed to get http host address: %s", err)
	}

	client, err := api.NewClient(&api.Config{
		Address: httpHostAddress,
	})
	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}
	client.SetToken("unittesttoken")

	// put secrets
	secretPath := "secret/data/secret"
	secret := map[string]interface{}{"data": map[string]interface{}{"key1": "value1", "key2": "value2", "key3": "value3", "key4": 123, "key5": true}}

	_, err = client.Logical().Write(secretPath, secret)
	if err != nil {
		t.Fatalf("failed to write secret: %s", err)
	}

	testClient = client
}

// TestExtractSecretsWithFormatAsExpected tests if a two secrets one with a format is extracted correct
func TestExtractSecretsWithFormatAsExpected(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	// define input
	data := files.Read("../test_data/two_secrets_with_format.yaml")
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
	for _, v := range result {

		// assert
		expected := fmt.Sprintf("%v", map[string]interface{}{input.Prefix + "key1": "value1", input.Prefix + "key2": "value2", input.Prefix + "key3": "value3", input.Prefix + "key4": float64(123), input.Prefix + "key5": true})
		actual := fmt.Sprintf("%v", v.Result)

		assert.Equal(t, expected, actual)
	}

}

// TestExtractSecretsAsExpected tests if a simple secret is extracted correct
func TestExtractSecretsAsExpected(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	// define input
	data := files.Read("../test_data/single_secret.yaml")
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
	for _, v := range result {
		// assert
		expected := fmt.Sprintf("%v", map[string]interface{}{input.Prefix + "key1": "value1", input.Prefix + "key2": "value2", input.Prefix + "key3": "value3", input.Prefix + "key4": float64(123), input.Prefix + "key5": true})
		actual := fmt.Sprintf("%v", v.Result)

		assert.Equal(t, expected, actual)
	}

}

// TestExtractSecretsWithPrefixAsExpected tests if a simple secret is extracted correct
func TestExtractSecretsWithPrefixAsExpected(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	// define input
	data := files.Read("../test_data/keys_with_prefix.yaml")
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

	for _, v := range result {
		// assert
		expected := fmt.Sprintf("%v", map[string]interface{}{"PRE_key1": "value1", "FIX_key2": "value2", input.Prefix + "key3": "value3"})
		actual := fmt.Sprintf("%v", v.Result)

		assert.Equal(t, expected, actual)
	}

}

// TestExtractSecretsSaveAsFileAsExpected tests if a simple secret is extracted correct
func TestExtractSecretsSaveAsFileAsExpected(t *testing.T) {
	// arrange
	setupVault(t)
	t.Cleanup(func() {
		testClient = nil
	})

	// define input
	data := files.Read("../test_data/save_as_file.yaml")
	input := util.ReadInput(data)

	// mock prefix
	config.Config.Prefix = input.Prefix

	vaultClient := &API{
		Client: testClient,
	}

	// act
	_, err := vaultClient.ExtractSecrets(input, false)
	if err != nil {
		t.Error(err)
	}

	// clean up
	defer os.Remove("../.tmp/TEST_key1")

	// assert
	content, err := os.ReadFile("../.tmp/TEST_key1")
	if err != nil {
		t.Errorf("could not read file: %v", err)
	}

	expected := "value1"
	actual := string(content)

	assert.Equal(t, expected, actual)

}
