package main

import (
	"os"

	"github.com/BESTSELLER/harpocrates/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
