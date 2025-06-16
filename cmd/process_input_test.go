package cmd

import (
	// "fmt" // Removed unused import
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/util" // Assuming SecretJSON is here
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a dummy root command and initialize flags
func newTestRootCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "harpocrates"}
	config.InitFlags(cmd) // Initialize flags as in main application
	return cmd
}

func TestProcessInput_File(t *testing.T) {
	validYAMLContent := `
secrets:
  - secret/data/path1:
      keys: ["key1"]
`
	// invalidYAMLContent := `
	// secrets:
	//   - secret/data/path1:
	//       keys: ["key1"]
	// ` // This is actually valid YAML, schema validation is what would fail.
	// For this test, we'll use a structurally invalid secrets file content per schema
	// (e.g. 'keys' is a string instead of list)
	// This variable is currently unused. Removing it.
	structurallyInvalidContent := `
secrets:
  - secret/data/path1:
      keys: "not-a-list"
`

	malformedYAMLContent := `key: value: another_value`

	tempDir := t.TempDir()

	// --- Test Case 1: Valid file, valid content ---
	t.Run("valid file and content", func(t *testing.T) {
		cmd := newTestRootCmd()
		filePath := filepath.Join(tempDir, "valid.yaml")
		require.NoError(t, os.WriteFile(filePath, []byte(validYAMLContent), 0644))

		// Simulate setting the file flag
		cmd.Flags().Set("file", filePath)
		// Manually update the unexported variable that GetSecretFile reads.
		// This is a workaround. A better solution would be an interface or exported setter.
		// For now, we rely on how InitFlags sets the global `secretFile` in config package.
		// This requires `secretFile` in `config` to be accessible or `GetSecretFile` to be mockable.
		// Since `config.secretFile` is not exported, we'll test by ensuring cobra sets it.
		// Cobra sets the variable bound by StringVarP, which is `config.secretFile` via `config.InitFlags`.
		// So, parsing the flag should update it.

		// We need to parse the flags for `config.GetSecretFile` to work correctly.
		args := []string{"--file", filePath}
		require.NoError(t, cmd.ParseFlags(args)) // This will set the `secretFile` var in config pkg

		config.Config.Validate = false // Ensure validate mode is off

		input, err := processInput(cmd, []string{}) // args for processInput are for inline secrets
		assert.NoError(t, err)
		assert.NotEmpty(t, input.Secrets)
		assert.Equal(t, "secret/data/path1", getFirstSecretPath(input))
	})

	// --- Test Case 2: File read error (file does not exist) ---
	t.Run("file read error", func(t *testing.T) {
		cmd := newTestRootCmd()
		nonExistentFilePath := filepath.Join(tempDir, "nonexistent.yaml")
		args := []string{"--file", nonExistentFilePath}
		require.NoError(t, cmd.ParseFlags(args))

		config.Config.Validate = false
		_, err := processInput(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read secret file")
	})

	// --- Test Case 3: Valid file, but content fails schema validation ---
	t.Run("file content schema validation error", func(t *testing.T) {
		cmd := newTestRootCmd()
		filePath := filepath.Join(tempDir, "invalid_schema.yaml")
		require.NoError(t, os.WriteFile(filePath, []byte(structurallyInvalidContent), 0644))
		args := []string{"--file", filePath}
		require.NoError(t, cmd.ParseFlags(args))

		config.Config.Validate = false // Test actual processing
		_, err := processInput(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret file validation failed")
	})

	// --- Test Case 4: Malformed YAML content ---
	t.Run("malformed yaml content", func(t *testing.T) {
		cmd := newTestRootCmd()
		filePath := filepath.Join(tempDir, "malformed.yaml")
		require.NoError(t, os.WriteFile(filePath, []byte(malformedYAMLContent), 0644))
		args := []string{"--file", filePath}
		require.NoError(t, cmd.ParseFlags(args))

		config.Config.Validate = false
		_, err := processInput(cmd, []string{})
		assert.Error(t, err)
		// The error comes from yaml.YAMLToJSON via validate.SecretsFile
		assert.Contains(t, err.Error(), "failed to convert YAML to JSON")
	})

	// --- Test Case 5: --validate flag with valid file ---
	t.Run("validate flag with valid file", func(t *testing.T) {
		cmd := newTestRootCmd()
		filePath := filepath.Join(tempDir, "validate_valid.yaml")
		require.NoError(t, os.WriteFile(filePath, []byte(validYAMLContent), 0644))
		args := []string{"--file", filePath, "--validate"} // Simulate --validate flag

		// Parse flags to set config.Config.Validate and config.secretFile
		require.NoError(t, cmd.ParseFlags(args))
		// Ensure config.Config.Validate is true after parsing
		require.True(t, config.Config.Validate, "config.Config.Validate should be true")

		_, err := processInput(cmd, []string{})
		assert.Equal(t, ErrValidationSuccessful, err)
	})
}

func TestProcessInput_SecretSlice(t *testing.T) {
	// --- Test Case 1: Valid secret slice ---
	t.Run("valid secret slice", func(t *testing.T) {
		cmd := newTestRootCmd()
		// Simulate setting the secret slice flag
		secrets := []string{"secret/data/slicepath1", "secret/data/slicepath2"}
		secretsStr := strings.Join(secrets, ",")
		args := []string{"--secret", secretsStr} // Cobra string slice parsing
		require.NoError(t, cmd.ParseFlags(args))

		config.Config.Output = "/tmp/harpocrates_test_output" // Required for slice processing
		config.Config.Validate = false
		defer func() { config.Config.Output = "" }()

		input, err := processInput(cmd, []string{})
		assert.NoError(t, err)
		assert.Len(t, input.Secrets, 2)
		assert.Equal(t, "secret/data/slicepath1", input.Secrets[0].(string))
	})

	// --- Test Case 2: Secret slice without output specified ---
	t.Run("secret slice no output", func(t *testing.T) {
		cmd := newTestRootCmd()
		secrets := []string{"secret/data/slicepath1"}
		secretsStr := strings.Join(secrets, ",")
		args := []string{"--secret", secretsStr}
		require.NoError(t, cmd.ParseFlags(args))

		originalOutput := config.Config.Output
		config.Config.Output = "" // Ensure output is not set
		config.Config.Validate = false
		defer func() { config.Config.Output = originalOutput }()

		_, err := processInput(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "output is required when using secret parameters")
	})
}

func TestProcessInput_Inline(t *testing.T) {
	validInlineYAML := `secrets: [{ "secret/data/inlinepath": { keys: ["keyA"] } }]`
	invalidInlineContent := `secrets: [{ "secret/data/inlinepath": { keys: "not-a-list" } }]`

	// --- Test Case 1: Valid inline content ---
	t.Run("valid inline content", func(t *testing.T) {
		cmd := newTestRootCmd()
		config.Config.Validate = false

		input, err := processInput(cmd, []string{validInlineYAML})
		assert.NoError(t, err)
		assert.NotEmpty(t, input.Secrets)
		assert.Equal(t, "secret/data/inlinepath", getFirstSecretPath(input))
	})

	// --- Test Case 2: Invalid inline content (schema validation) ---
	t.Run("invalid inline content schema", func(t *testing.T) {
		cmd := newTestRootCmd()
		config.Config.Validate = false

		_, err := processInput(cmd, []string{invalidInlineContent})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inline secret validation failed")
	})

	// --- Test Case 3: --validate flag with valid inline content ---
	t.Run("validate flag with valid inline", func(t *testing.T) {
		cmd := newTestRootCmd()
		args := []string{"--validate"}           // Simulate --validate flag
		require.NoError(t, cmd.ParseFlags(args)) // Sets config.Config.Validate
		require.True(t, config.Config.Validate)

		_, err := processInput(cmd, []string{validInlineYAML})
		assert.Equal(t, ErrValidationSuccessful, err)
	})

	// --- Test Case 4: No input provided (no file, no slice, no inline) ---
	t.Run("no input provided", func(t *testing.T) {
		cmd := newTestRootCmd()
		config.Config.Validate = false
		// Ensure no flags are set that would imply other input methods
		// GetSecretFile and GetSecretSlice should return empty due to no flags parsed for them

		_, err := processInput(cmd, []string{}) // No inline args either
		assert.Error(t, err)
		// This error now comes from the `else` block in processInput when len(args) == 0
		assert.Contains(t, err.Error(), "no input provided: use --file, --secret, or inline JSON")
	})
}

// Helper to get the first secret path for assertion, more robustly.
func getFirstSecretPath(input util.SecretJSON) string {
	if len(input.Secrets) == 0 {
		return ""
	}
	firstSecretEntry := input.Secrets[0]

	// Check if it's a simple string path
	if pathStr, ok := firstSecretEntry.(string); ok {
		return pathStr
	}

	// Check if it's a map (complex secret)
	if secretMap, ok := firstSecretEntry.(map[string]interface{}); ok {
		// In a complex secret, the structure is { "path/to/secret": { ...config... } }
		// So, we iterate over the single key in this map which is the path.
		for path := range secretMap {
			return path
		}
	}
	return ""
}

func TestMain(m *testing.M) {
	// This TestMain is intentionally left empty for now, as vault_test.go has its own.
	// If we need package-level setup for cmd tests, it would go here.
	// For now, individual tests manage their setup.
	os.Exit(m.Run())
}
