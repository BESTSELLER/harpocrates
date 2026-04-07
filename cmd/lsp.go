package cmd

import (
	"github.com/BESTSELLER/harpocrates/internal/lsp"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/rs/zerolog/log"
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
	err := loadLocalVaultToken()
	if err != nil {
		log.Warn().Err(err).Msg("Vault token validation failed, autocomplete/validation may not work")
	}
	vault.Login()
	vaultClient := vault.NewClient()

	server := lsp.NewServer(vaultClient, err)
	server.Start()
}
