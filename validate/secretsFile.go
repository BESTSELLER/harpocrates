package validate

import (
	// Used for embedding the schema
	_ "embed"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

//go:embed schema.json
var schema string

// SecretsFile validates the secrets file and returns true or false depending on the validation result.
// Outputs error message if validation fails including what the issue is.
// Debug message is logged if debug is true and validation succeeded.
func SecretsFile(fileToValidate string) bool {
	y, err := yaml.YAMLToJSON([]byte(fileToValidate))
	if err != nil {
		panic(err)
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(string(y))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		log.Debug().Msg("Secrets file validated successfully!")
		return true
	}

	logArr := zerolog.Arr()
	for _, desc := range result.Errors() {
		logArr.Str(desc.String())
	}
	log.Error().Array("validation_errors", logArr).Msg("Secrets file failed validation")
	return false

}
