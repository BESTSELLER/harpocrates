package cmd

import (
	"github.com/spf13/cobra"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run a command with secrets injected from Vault",
	Long: `Run a command with secrets injected from Vault.

The dev command fetches secrets from Vault and makes them available to a child process.
It creates temporary files for the secrets and cleans them up after the command has finished executing.
This is useful for local development, where you want to run an application with secrets from Vault without storing them in plaintext on your machine.`,
	Run: func(cmd *cobra.Command, args []string) {
		doIt(cmd, args, true)
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}
