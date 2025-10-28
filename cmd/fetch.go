package cmd

import (
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch secrets and dump them somewhere",
	Run: func(cmd *cobra.Command, args []string) {
		doIt(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
