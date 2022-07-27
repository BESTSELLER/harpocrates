package main

import (
	"fmt"
	"io/ioutil"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

func main() {

	documentFile, err := ioutil.ReadFile("./secret.yaml")
	if err != nil {
		panic(err)
	}

	y, err := yaml.YAMLToJSON(documentFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(y))

	schemaLoader := gojsonschema.NewReferenceLoader("file://./schema.json")
	documentLoader := gojsonschema.NewStringLoader(string(y))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
}
