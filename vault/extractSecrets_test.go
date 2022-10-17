package vault

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/hashicorp/vault/api"
	vaulthttp "github.com/hashicorp/vault/http"
	hashivault "github.com/hashicorp/vault/vault"
	"gotest.tools/assert"
)

var testVault vaultTest

// rootVaultToken is the Vault token used for tests
var rootVaultToken = "unittesttoken"

type vaultTest struct {
	Cluster *hashivault.TestCluster
	Client  *api.Client
}

func TestMain(t *testing.T) {
	testVault = GetTestVaultServer(t)
}

// GetTestVaultServer creates the test server
func GetTestVaultServer(t *testing.T) vaultTest {
	t.Helper()

	cluster := hashivault.NewTestCluster(t, &hashivault.CoreConfig{
		DevToken: rootVaultToken,
	}, &hashivault.TestClusterOptions{
		HandlerFunc: vaulthttp.Handler,
	})
	cluster.Start()

	core := cluster.Cores[0].Core
	hashivault.TestWaitActive(t, core)
	client := cluster.Cores[0].Client

	// put secrets
	secretPath := fmt.Sprintf("secret/data/secret")
	secret := map[string]interface{}{"key1": "value1", "key2": "value2", "key3": "value3"}

	_, err := client.Logical().Write(secretPath, secret)
	if err != nil {
		panic(err)
	}

	// return server
	return vaultTest{
		Cluster: cluster,
		Client:  client,
	}

}

// TestExtractSecretsWithFormatAsExpected tests if a two secrets one with a format is extracted correct
func TestExtractSecretsWithFormatAsExpected(t *testing.T) {
	// arrange

	// define input
	data := files.Read("../test_data/two_secrets_with_format.yaml")
	input := util.ReadInput(data)

	// mock prefix
	config.Config.Prefix = input.Prefix

	var vaultClient *API
	vaultClient = &API{
		Client: testVault.Client,
	}

	// act
	result, err := vaultClient.ExtractSecrets(input, false)
	if err != nil {
		t.Error(err)
	}
	for _, v := range result {

		// assert
		expected := fmt.Sprintf("%v", map[string]interface{}{input.Prefix + "key1": "value1", input.Prefix + "key2": "value2", input.Prefix + "key3": "value3"})
		actual := fmt.Sprintf("%v", v.Result)

		assert.Equal(t, expected, actual)
	}

}

// TestExtractSecretsAsExpected tests if a simple secret is extracted correct
func TestExtractSecretsAsExpected(t *testing.T) {
	// arrange

	// define input
	data := files.Read("../test_data/single_secret.yaml")
	input := util.ReadInput(data)

	// mock prefix
	config.Config.Prefix = input.Prefix

	var vaultClient *API
	vaultClient = &API{
		Client: testVault.Client,
	}

	// act
	result, err := vaultClient.ExtractSecrets(input, false)
	if err != nil {
		t.Error(err)
	}
	for _, v := range result {
		// assert
		expected := fmt.Sprintf("%v", map[string]interface{}{input.Prefix + "key1": "value1", input.Prefix + "key2": "value2", input.Prefix + "key3": "value3"})
		actual := fmt.Sprintf("%v", v.Result)

		assert.Equal(t, expected, actual)
	}

}

// TestExtractSecretsWithPrefixAsExpected tests if a simple secret is extracted correct
func TestExtractSecretsWithPrefixAsExpected(t *testing.T) {
	// arrange

	// define input
	data := files.Read("../test_data/keys_with_prefix.yaml")
	input := util.ReadInput(data)

	// mock prefix
	config.Config.Prefix = input.Prefix

	var vaultClient *API
	vaultClient = &API{
		Client: testVault.Client,
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

	// define input
	data := files.Read("../test_data/save_as_file.yaml")
	input := util.ReadInput(data)

	// mock prefix
	config.Config.Prefix = input.Prefix

	var vaultClient *API
	vaultClient = &API{
		Client: testVault.Client,
	}

	// act
	_, err := vaultClient.ExtractSecrets(input, false)
	if err != nil {
		t.Error(err)
	}

	// clean up
	defer os.Remove("../.tmp/TEST_key1")

	// assert
	content, err := ioutil.ReadFile("../.tmp/TEST_key1")
	if err != nil {
		t.Errorf("could not read file: %v", err)
	}

	expected := "value1"
	actual := string(content)

	assert.Equal(t, expected, actual)

}
