package cmd

import (
	"github.com/BESTSELLER/harpocrates/internal/lsp"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the LSP server for harpocrates secrets files",
	Run: func(cmd *cobra.Command, args []string) {
		startLSP()
	},
}

func init() {
	rootCmd.AddCommand(lspCmd)
}

func startLSP() {
	// Silence standard logging output for LSP standard I/O (we only log to a file or disable it)
	// For production, maybe write to /tmp/harpocrates-lsp.log

	// Try to get Vault initialized identical to `fetch`
	loadLocalVaultToken()
	vault.Login()
	vaultClient := vault.NewClient()

	server := lsp.NewServer(vaultClient)
	server.Start()
}
