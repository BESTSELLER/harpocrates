package main

import "github.com/BESTSELLER/harpocrates/cmd"

var secretJSON string

func main() {
	cmd.Execute()
	// fmt.Println("Harpocrates has started...")
	// config.LoadConfig()
	// util.GetVaultToken()

	// args := os.Args[1:]

	// if len(args) == 0 {
	// 	fmt.Println("No secret file provided!")
	// 	os.Exit(1)
	// }
	// arg := args[0]

	// input := util.ReadInput(arg)
	// allSecrets := util.ExtractSecrets(input)

	// if input.Format == "json" {
	// 	files.WriteFile(input.Output, fmt.Sprintf("secrets.%s", input.Format), files.FormatAsJSON(allSecrets))
	// }

	// if input.Format == "env" {
	// 	files.WriteFile(input.Output, fmt.Sprintf("secrets.%s", input.Format), files.FormatAsENV(allSecrets))
	// }

}
