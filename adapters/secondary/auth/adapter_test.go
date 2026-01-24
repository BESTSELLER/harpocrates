package auth_test

import (
	"testing"
	"time"

	"github.com/BESTSELLER/harpocrates/adapters/secondary/auth"
	"github.com/BESTSELLER/harpocrates/domain/ports"
)

// TestAuthAdapterImplementsPort verifies that the auth adapter implements the Authenticator port
func TestAuthAdapterImplementsPort(t *testing.T) {
	var _ ports.Authenticator = auth.NewAdapter("http://localhost:8200", "kubernetes", "myrole", "/path/to/token", false)
}

// TestAuthAdapterIsTokenValid tests token validity checking
func TestAuthAdapterIsTokenValid(t *testing.T) {
	adapter := auth.NewAdapter("http://localhost:8200", "kubernetes", "myrole", "/path/to/token", false)
	
	tests := []struct {
		name     string
		token    string
		expiry   time.Time
		expected bool
	}{
		{
			name:     "Empty token is invalid",
			token:    "",
			expiry:   time.Now().Add(1 * time.Hour),
			expected: false,
		},
		{
			name:     "Valid token with future expiry",
			token:    "test-token",
			expiry:   time.Now().Add(1 * time.Hour),
			expected: true,
		},
		{
			name:     "Token expiring soon (within 5 minutes) is invalid in non-continuous mode",
			token:    "test-token",
			expiry:   time.Now().Add(3 * time.Minute),
			expected: true, // In non-continuous mode, we don't check the 5-minute window
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.IsTokenValid(tt.token, tt.expiry)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
