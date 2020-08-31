package util

import (
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/files"

	"github.com/BESTSELLER/harpocrates/vault"
)

// ExtractSecrets will loop through al those damn interfaces
func ExtractSecrets(input SecretJSON) map[string]interface{} {
	result := make(map[string]interface{})

	for _, a := range input.Secrets {

		// If the key is just a secret path, then it will read that from Vault, otherwise:
		if fmt.Sprintf("%T", a) != "string" {
			b, ok := a.(map[string]interface{})
			if !ok {
				fmt.Printf("Expected map[string]interface{}, got: '%s'\n", a)
				os.Exit(1)
			}

			for c, d := range b {
				e, ok := d.([]interface{})
				if !ok {
					fmt.Printf("Expected an array, got: '%s'\n", d)
					os.Exit(1)
				}
				for _, f := range e {

					// If the key is just a secret path, then it will read that from Vault, otherwise:
					if fmt.Sprintf("%T", f) != "string" {
						g, ok := f.(map[string]interface{})
						if !ok {
							fmt.Printf("Expected map[string]interface{}, got: '%s'\n", a)
							os.Exit(1)
						}
						for h, i := range g {
							j, ok := i.(map[string]interface{})
							if !ok {
								fmt.Printf("Expected map[string]interface{}, got: '%s'\n", a)
								os.Exit(1)
							}

							some := j["saveAsFile"]
							if some != nil {
								var secretValue = vault.ReadSecretKey(fmt.Sprintf("%s", c), h)
								if some.(bool) {
									fmt.Println("Creating file...", h)
									files.WriteFile(input.DirPath, h, secretValue)
								} else {
									result[h] = secretValue
								}
							}
						}
					} else {
						var secretValue = vault.ReadSecretKey(fmt.Sprintf("%s", c), fmt.Sprintf("%s", f))
						result[fmt.Sprintf("%s", f)] = secretValue
					}
				}
			}
		} else {
			var secretValue = vault.ReadSecret(fmt.Sprintf("%s", a))
			for aa, bb := range secretValue {
				result[aa] = bb
			}
		}
	}
	return result
}
