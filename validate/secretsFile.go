package validate

import (
	_ "embed"
	"fmt" // Correctly placed import

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

//go:embed schema.json
var schema string

// ErrValidationFailed is returned when JSON schema validation fails.
type ErrValidationFailed struct {
	ValidationErrors []gojsonschema.ResultError
}

func (e *ErrValidationFailed) Error() string {
	var errStrings []string
	for _, desc := range e.ValidationErrors {
		errStrings = append(errStrings, desc.String())
	}
	return fmt.Sprintf("secrets file validation failed: %v", errStrings)
}

// SecretsFile validates the structure of the secrets file against a JSON schema.
// It returns nil if the file is valid, or an error otherwise.
// If validation fails, the error will be of type *ErrValidationFailed.
func SecretsFile(fileToValidate string) error {
	jsonData, err := yaml.YAMLToJSON([]byte(fileToValidate))
	if err != nil {
		return fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(string(jsonData))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		// This error is from the validation process itself, not schema validation failure
		return fmt.Errorf("error during schema validation process: %w", err)
	}

	if result.Valid() {
		log.Debug().Msg("Secrets file validated successfully!")
		return nil
	}

	// Validation failed, collect errors
	validationErrors := result.Errors()
	logArr := zerolog.Arr()
	for _, desc := range validationErrors {
		logArr.Str(desc.String())
	}
	log.Error().Array("validation_errors", logArr).Msg("Secrets file failed validation")

	return &ErrValidationFailed{ValidationErrors: validationErrors}
}
