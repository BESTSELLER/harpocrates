package vault

import (
	"testing"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretContext(t *testing.T) {
	// Setup global config for the test
	originalConfig := config.Config
	config.Config.Prefix = "global_prefix_"
	config.Config.UpperCase = true
	config.Config.Format = "global_format"
	defer func() { config.Config = originalConfig }()

	ctx := newSecretContext()

	assert.Equal(t, "global_prefix_", ctx.prefix)
	assert.True(t, ctx.upperCase)
	assert.Equal(t, "global_format", ctx.format)
}

func TestSecretContext_Update(t *testing.T) {
	// Setup global config for defaults
	originalConfig := config.Config
	config.Config.Prefix = "global_prefix_"
	config.Config.UpperCase = false // Global default for uppercase
	config.Config.Format = "global_format"
	defer func() { config.Config = originalConfig }()

	baseCtx := newSecretContext() // prefix="global_prefix_", upperCase=false, format="global_format"

	trueVal := true
	// falseVal := false // Removed unused variable

	tests := []struct {
		name                  string
		prefixOverride        string
		upperCaseOverride     *bool
		formatOverride        string
		resetToGlobalDefaults bool // Explicitly control this for each test case
		expectedPrefix        string
		expectedUpperCase     bool
		expectedFormat        string
	}{
		{
			name:                  "all overrides present, no reset",
			prefixOverride:        "override_prefix_",
			upperCaseOverride:     &trueVal,
			formatOverride:        "override_format",
			resetToGlobalDefaults: false,
			expectedPrefix:        "override_prefix_",
			expectedUpperCase:     true,
			expectedFormat:        "override_format",
		},
		{
			name:                  "empty overrides, with resetToGlobalDefaults=true (should use global)",
			prefixOverride:        "",
			upperCaseOverride:     nil,
			formatOverride:        "",
			resetToGlobalDefaults: true,
			expectedPrefix:        "global_prefix_", // Global
			expectedUpperCase:     false,            // Global
			expectedFormat:        "global_format",  // Global
		},
		{
			name:                  "empty overrides, with resetToGlobalDefaults=false (should inherit from baseCtx)",
			prefixOverride:        "",                // Empty, should inherit
			upperCaseOverride:     nil,               // Nil, should inherit
			formatOverride:        "",                // Empty, should inherit
			resetToGlobalDefaults: false,             // Key: do not reset to global
			expectedPrefix:        baseCtx.prefix,    // Inherited
			expectedUpperCase:     baseCtx.upperCase, // Inherited
			expectedFormat:        baseCtx.format,    // Inherited
		},
		{
			name:                  "only prefix override, no reset (others inherit)",
			prefixOverride:        "specific_prefix_",
			upperCaseOverride:     nil,
			formatOverride:        "",
			resetToGlobalDefaults: false,
			expectedPrefix:        "specific_prefix_",
			expectedUpperCase:     baseCtx.upperCase, // Inherit
			expectedFormat:        baseCtx.format,    // Inherit
		},
		{
			name:                  "only uppercase override to true, no reset (others inherit)",
			prefixOverride:        "",
			upperCaseOverride:     &trueVal,
			formatOverride:        "",
			resetToGlobalDefaults: false,
			expectedPrefix:        baseCtx.prefix, // Inherit
			expectedUpperCase:     true,
			expectedFormat:        baseCtx.format, // Inherit
		},
		{
			name:                  "prefix override, others empty with reset=true (should reset others to global)",
			prefixOverride:        "specific_prefix_reset_others",
			upperCaseOverride:     nil, // Will be reset to global
			formatOverride:        "",  // Will be reset to global
			resetToGlobalDefaults: true,
			expectedPrefix:        "specific_prefix_reset_others",
			expectedUpperCase:     false,           // Global default
			expectedFormat:        "global_format", // Global default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedCtx := baseCtx.update(tt.prefixOverride, tt.upperCaseOverride, tt.formatOverride, tt.resetToGlobalDefaults)
			assert.Equal(t, tt.expectedPrefix, updatedCtx.prefix, "Prefix mismatch")
			assert.Equal(t, tt.expectedUpperCase, updatedCtx.upperCase, "UpperCase mismatch")
			assert.Equal(t, tt.expectedFormat, updatedCtx.format, "Format mismatch")

			// Ensure baseCtx was not modified (update should return a new context)
			assert.Equal(t, "global_prefix_", baseCtx.prefix, "Base context prefix should not change")
			assert.Equal(t, false, baseCtx.upperCase, "Base context upperCase should not change")
			assert.Equal(t, "global_format", baseCtx.format, "Base context format should not change")
		})
	}
}
