package main

import (
	"fmt"
	"os"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
)

var secretJSON string

func main() {
	fmt.Println("Harpocrates has started...")
	config.LoadConfig()
	util.GetVaultToken()

	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("No secret file provided!")
		os.Exit(1)
	}
	arg := args[0]

	input := util.ReadInput(arg)
	allSecrets := util.ExtractSecrets(input)

	if input.Format == "json" {
		files.WriteFile(input.DirPath, fmt.Sprintf("secrets.%s", input.Format), files.FormatAsJSON(allSecrets))
	}

	if input.Format == "env" {
		files.WriteFile(input.DirPath, fmt.Sprintf("secrets.%s", input.Format), files.FormatAsENV(allSecrets))
	}

}
