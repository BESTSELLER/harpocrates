package cmd

import (
	"os"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/internal/lsp"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/BESTSELLER/harpocrates/vault/gcp"
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

	if config.Config.GcpWorkloadID {
		login, err := gcp.FetchVaultLogin(config.Config.VaultAddress, config.Config.AuthName)
		if err == nil {
			config.Config.VaultToken = login.Auth.ClientToken
		}
	}

	var vaultClient *vault.API
	// Try establishing the vault client if auth was loaded, if not gracefully fall back.
	// We do not fail hard in LSP!
	if os.Getenv("VAULT_TOKEN") != "" || os.Getenv("VAULT_ADDR") != "" {
		vaultClient = vault.NewClient()
	}

	server := lsp.NewServer(vaultClient)
	server.Start()
}
