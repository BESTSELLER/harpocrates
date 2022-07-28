package validate

import (
	"fmt"
	"io/ioutil"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

// Validate secrets file and returns true or false depending on the validation result.
// Outputs error message if validation fails including what the issue is.
// Debug message is logged if debug is true and validation succeeded.
func SecretsFile(f string) bool {
	documentFile, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	y, err := yaml.YAMLToJSON(documentFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(y))

	schemaLoader := gojsonschema.NewReferenceLoader("file://./validate/schema.json")
	documentLoader := gojsonschema.NewStringLoader(string(y))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		log.Debug().Msg("Secrets file validated successfully!")
		return true
	} else {
		logArr := zerolog.Arr()
		for _, desc := range result.Errors() {
			logArr.Str(desc.String())
		}
		log.Error().Array("validation_errors", logArr).Msg("Secrets file failed validation")
		return false
	}
}