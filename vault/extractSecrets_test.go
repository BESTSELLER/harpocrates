package vault

import (
	"fmt"
	"testing"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/hashicorp/vault/api"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"gotest.tools/assert"
)

const (
	// rootVaultToken is the Vault token used for tests
	rootVaultToken = "unittesttoken"
)

type vaultTest struct {
	Cluster *vault.TestCluster
	Client  *api.Client
}

// creates the test server
func GetTestVaultServer(t *testing.T) vaultTest {
	t.Helper()

	cluster := vault.NewTestCluster(t, &vault.CoreConfig{
		DevToken: rootVaultToken,
	}, &vault.TestClusterOptions{
		HandlerFunc: vaulthttp.Handler,
	})
	cluster.Start()

	core := cluster.Cores[0].Core
	vault.TestWaitActive(t, core)
	client := cluster.Cores[0].Client

	return vaultTest{
		Cluster: cluster,
		Client:  client,
	}
}

// TestExtractSecretsAsExpected tests if the config is loaded as expected
func TestExtractSecretsAsExpected(t *testing.T) {
	// arrange

	// spin up vault and put secret
	testVault := GetTestVaultServer(t)
	defer testVault.Cluster.Cleanup()

	secretPath := fmt.Sprintf("secret/data/secret")
	secret := map[string]interface{}{"foo": "bar"}

	_, err := testVault.Client.Logical().Write(secretPath, secret)
	if err != nil {
		panic(err)
	}

	// define input
	var secrets []interface{}
	secrets = append(secrets, secretPath)

	input := util.SecretJSON{
		Format:  "env",
		Prefix:  "TEST_",
		Output:  "pbtest",
		Secrets: secrets,
	}

	// mock prefix
	config.Config.Prefix = input.Prefix

	var vaultClient *API
	vaultClient = &API{
		Client: testVault.Client,
	}

	// act
	result := vaultClient.ExtractSecrets(input)

	// assert
	expected := fmt.Sprintf("%v", map[string]interface{}{input.Prefix + "foo": "bar"})
	actual := fmt.Sprintf("%v", result)

	assert.Equal(t, expected, actual)

}
