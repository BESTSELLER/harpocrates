package lsp

type VaultClient interface {
	ListKeys(path string) ([]string, error)
	ListSecretEngines() ([]string, error)
	GetEngineSubPath(mountPath string) (string, error)
	ReadSecret(path string) (map[string]any, error)
}
