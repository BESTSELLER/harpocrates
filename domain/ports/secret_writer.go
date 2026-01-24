package ports

// SecretWriter defines the port for writing secrets to storage
type SecretWriter interface {
	// Write writes content to a file at the specified output path
	Write(output string, fileName string, content interface{}, owner *int, append bool) error
	
	// Read reads the content of a file
	Read(filePath string) (string, error)
}
