package ports

// SecretFetcher defines the port for fetching secrets from a secret store
type SecretFetcher interface {
	// ReadSecret reads all key-value pairs from a secret path
	ReadSecret(path string) (map[string]interface{}, error)
	
	// ReadSecretKey reads a specific key from a secret path
	ReadSecretKey(path string, key string) (string, error)
}
