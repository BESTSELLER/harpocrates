package files_test

import (
	"testing"

	"github.com/BESTSELLER/harpocrates/domain/ports"
	"github.com/BESTSELLER/harpocrates/files"
)

// TestSetAdapter demonstrates that the files package can now be tested with a mock adapter
func TestSetAdapter(t *testing.T) {
	// Create a mock adapter
	mockAdapter := &mockWriter{
		readContent: "test content",
	}
	
	// Inject the mock
	files.SetAdapter(mockAdapter)
	defer files.ResetAdapter()
	
	// Test Read
	content := files.Read("/test/path")
	if content != "test content" {
		t.Errorf("Expected 'test content', got '%s'", content)
	}
	
	if len(mockAdapter.readCalls) != 1 {
		t.Errorf("Expected 1 read call, got %d", len(mockAdapter.readCalls))
	}
	
	// Test Write
	files.Write("/tmp", "test.txt", "data", nil, false)
	
	if len(mockAdapter.writeCalls) != 1 {
		t.Errorf("Expected 1 write call, got %d", len(mockAdapter.writeCalls))
	}
	
	if mockAdapter.writeCalls[0].fileName != "test.txt" {
		t.Errorf("Expected fileName 'test.txt', got '%s'", mockAdapter.writeCalls[0].fileName)
	}
}

// Mock adapter for testing
type mockWriter struct {
	readContent string
	readCalls   []string
	writeCalls  []writeCall
}

type writeCall struct {
	output   string
	fileName string
	content  interface{}
}

func (m *mockWriter) Read(filePath string) (string, error) {
	m.readCalls = append(m.readCalls, filePath)
	return m.readContent, nil
}

func (m *mockWriter) Write(output string, fileName string, content interface{}, owner *int, appendToFile bool) error {
	m.writeCalls = append(m.writeCalls, writeCall{
		output:   output,
		fileName: fileName,
		content:  content,
	})
	return nil
}

var _ ports.SecretWriter = (*mockWriter)(nil)
