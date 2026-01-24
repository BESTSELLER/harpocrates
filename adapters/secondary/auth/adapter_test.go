package auth_test

import (
	"testing"
	"time"

	"github.com/BESTSELLER/harpocrates/adapters/secondary/auth"
	"github.com/BESTSELLER/harpocrates/domain/ports"
)

// TestAuthAdapterImplementsPort verifies that the auth adapter implements the Authenticator port
func TestAuthAdapterImplementsPort(t *testing.T) {
	var _ ports.Authenticator = auth.NewAdapter("http://localhost:8200", "kubernetes", "myrole", "/path/to/token", false, false)
}

// TestAuthAdapterIsTokenValid tests token validity checking
func TestAuthAdapterIsTokenValid(t *testing.T) {
	adapterNonContinuous := auth.NewAdapter("http://localhost:8200", "kubernetes", "myrole", "/path/to/token", false, false)
	adapterContinuous := auth.NewAdapter("http://localhost:8200", "kubernetes", "myrole", "/path/to/token", false, true)
	
	tests := []struct {
		name        string
		adapter     ports.Authenticator
		token       string
		expiry      time.Time
		expected    bool
		description string
	}{
		{
			name:        "Empty token is invalid",
			adapter:     adapterNonContinuous,
			token:       "",
			expiry:      time.Now().Add(1 * time.Hour),
			expected:    false,
			description: "Empty token should always be invalid",
		},
		{
			name:        "Valid token with future expiry in non-continuous mode",
			adapter:     adapterNonContinuous,
			token:       "test-token",
			expiry:      time.Now().Add(1 * time.Hour),
			expected:    true,
			description: "Valid token with distant expiry in non-continuous mode",
		},
		{
			name:        "Token expiring soon in non-continuous mode is still valid",
			adapter:     adapterNonContinuous,
			token:       "test-token",
			expiry:      time.Now().Add(3 * time.Minute),
			expected:    true,
			description: "In non-continuous mode, we don't check the 5-minute window",
		},
		{
			name:        "Token expiring soon in continuous mode is invalid",
			adapter:     adapterContinuous,
			token:       "test-token",
			expiry:      time.Now().Add(3 * time.Minute),
			expected:    false,
			description: "In continuous mode, tokens expiring within 5 minutes are invalid",
		},
		{
			name:        "Valid token with future expiry in continuous mode",
			adapter:     adapterContinuous,
			token:       "test-token",
			expiry:      time.Now().Add(1 * time.Hour),
			expected:    true,
			description: "Valid token with distant expiry in continuous mode",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adapter.IsTokenValid(tt.token, tt.expiry)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v. %s", tt.expected, result, tt.description)
			}
		})
	}
}
