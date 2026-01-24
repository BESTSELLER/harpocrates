package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BESTSELLER/harpocrates/adapters/secondary/filesystem"
	"github.com/BESTSELLER/harpocrates/domain/ports"
)

// TestFilesystemAdapterImplementsPort verifies that the filesystem adapter implements the SecretWriter port
func TestFilesystemAdapterImplementsPort(t *testing.T) {
	var _ ports.SecretWriter = filesystem.NewAdapter()
}

// TestFilesystemAdapterWriteRead tests write and read operations
func TestFilesystemAdapterWriteRead(t *testing.T) {
	adapter := filesystem.NewAdapter()
	
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "harpocrates-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Test write
	content := "test-secret-value"
	fileName := "test.txt"
	err = adapter.Write(tmpDir, fileName, content, nil, false)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	
	// Test read
	filePath := filepath.Join(tmpDir, fileName)
	readContent, err := adapter.Read(filePath)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	
	if readContent != content {
		t.Errorf("Expected content %s, got %s", content, readContent)
	}
}
