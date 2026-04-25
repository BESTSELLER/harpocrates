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

	keysInterface, ok := secretValues.Data["keys"].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected type for keys at path '%s'", path)
	}

	var keys []string
	for _, k := range keysInterface {
		keys = append(keys, k.(string))
	}

	return keys, nil
}

// ListSecretEngines lists all secret engines (mounts) available at the root level.
func (client *API) ListSecretEngines() ([]string, error) {
	mounts, err := client.Client.Sys().ListMounts()
	if err != nil {
		return nil, fmt.Errorf("failed to list secret engines: %v", err)
	}

	var engines []string
	for path := range mounts {
		if path != "" {
			engines = append(engines, path)
		}
	}

	return engines, nil
}

// GetEngineSubPath suggests the next path component (e.g., "data/", "roleset/") based on the engine type.
func (client *API) GetEngineSubPath(mountPath string) (string, error) {
	mounts, err := client.Client.Sys().ListMounts()
	if err != nil {
		return "", fmt.Errorf("failed to list mounts: %v", err)
	}

	mount, ok := mounts[mountPath]
	if !ok {
		return "", nil // Not a known root mount point
	}

	switch mount.Type {
	case "kv":
		if mount.Options != nil && mount.Options["version"] == "2" {
			return mountPath + "data/", nil
		}
	case "gcp":
		return mountPath + "roleset/", nil
	case "pki", "ssh", "aws", "database":
		return mountPath + "roles/", nil
	case "transit":
		return mountPath + "keys/", nil
	}

	return "", nil
}
