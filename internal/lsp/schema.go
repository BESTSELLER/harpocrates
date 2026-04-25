package lsp

import (
	"encoding/json"
	"sort"

	"github.com/BESTSELLER/harpocrates/validate"
)

var (
	rootFields     []string
	secretFields   []string
	keyFields      []string
	rootFieldVals  map[string][]string
	secretFieldVals map[string][]string
	keyFieldVals   map[string][]string
)

func init() {
	parseSchema()
}

func parseSchema() {
	var schemaRoot map[string]any
	if err := json.Unmarshal([]byte(validate.Schema), &schemaRoot); err != nil {
		panic(err)
	}

	// Initialize maps
	rootFieldVals = make(map[string][]string)
	secretFieldVals = make(map[string][]string)
	keyFieldVals = make(map[string][]string)

	// 1. Root fields
	if props, ok := schemaRoot["properties"].(map[string]any); ok {
		rootFields = extractKeys(props)
		extractFieldValues(props, rootFieldVals)
	}

	// Navigate to patternProperties of the secretsObject
	defs, _ := schemaRoot["$defs"].(map[string]any)
	secretsObj, _ := defs["secretsObject"].(map[string]any)
	patternProps, _ := secretsObj["patternProperties"].(map[string]any)

	// The key is the regex "^(?:(\\S| )+\\/)*(\\S| )+$"
	var secretPattern map[string]any
	for _, v := range patternProps {
		secretPattern, _ = v.(map[string]any)
		break // just take the first/only pattern
	}

	// 2. Secret fields
	if secProps, ok := secretPattern["properties"].(map[string]any); ok {
		secretFields = extractKeys(secProps)
		extractFieldValues(secProps, secretFieldVals)

		// Navigate to key fields
		if keysProp, ok := secProps["keys"].(map[string]any); ok {
			items, _ := keysProp["items"].(map[string]any)
			anyOf, _ := items["anyOf"].([]any)
			if len(anyOf) > 1 {
				keyObj, _ := anyOf[1].(map[string]any)
				keyPatternProps, _ := keyObj["patternProperties"].(map[string]any)

				var keyPattern map[string]any
				for _, v := range keyPatternProps {
					keyPattern, _ = v.(map[string]any)
					break
				}

				// 3. Key fields
				if keyProps, ok := keyPattern["properties"].(map[string]any); ok {
					keyFields = extractKeys(keyProps)
					extractFieldValues(keyProps, keyFieldVals)
				}
			}
		}
	}
}

func extractKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func extractFieldValues(props map[string]any, vals map[string][]string) {
	for fieldName, fieldDef := range props {
		if fieldObj, ok := fieldDef.(map[string]any); ok {
			// Check for enum
			if enumVals, ok := fieldObj["enum"].([]any); ok {
				values := make([]string, 0, len(enumVals))
				for _, v := range enumVals {
					if str, ok := v.(string); ok {
						values = append(values, str)
					}
				}
				if len(values) > 0 {
					sort.Strings(values)
					vals[fieldName] = values
				}
			}
			// Check for boolean type
			if typeVal, ok := fieldObj["type"].(string); ok && typeVal == "boolean" {
				vals[fieldName] = []string{"true", "false"}
			}
		}
	}
}

func GetRootFields() []string {
	return rootFields
}

func GetSecretFields() []string {
	return secretFields
}

func GetKeyFields() []string {
	return keyFields
}

func GetRootFieldVals() map[string][]string {
	return rootFieldVals
}

func GetSecretFieldVals() map[string][]string {
	return secretFieldVals
}

func GetKeyFieldVals() map[string][]string {
	return keyFieldVals
}
