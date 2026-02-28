package vault

import (
	"fmt"
)

// ListTokens lists the sub-paths or secrets at a specific Vault path.
func (client *API) ListTokens(path string) ([]string, error) {
	secretValues, err := client.Client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys at path '%s': %v", path, err)
	}

	if secretValues == nil || secretValues.Data == nil {
		return nil, nil
	}

	keysInterface, ok := secretValues.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for keys at path '%s'", path)
	}

	var keys []string
	for _, k := range keysInterface {
		keys = append(keys, k.(string))
	}

	return keys, nil
}
