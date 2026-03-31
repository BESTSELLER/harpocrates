package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/rs/zerolog/log"
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
		// Create temporary directory and file
		dir, err := os.MkdirTemp("", "harpocrates")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create temporary directory")
		}
		defer os.RemoveAll(dir) //nolint:errcheck // Best-effort cleanup of the temp folder; ignore errors as failure is non-critical and the OS will eventually reclaim the space.

		config.Config.Output = path.Join(dir, config.Config.Output)

		fmt.Println("output:", config.Config.Output)

		secretEnvs := doIt(cmd, args)

		// Set up cancellable context and signal handling for ctrl+c
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-signals
			cancel()
		}()

		fmt.Println(config.Config.Output)
		// Start the child application with the temporary file path using the context
		// cmd := exec.CommandContext(ctx, "bash", "-c", "echo $HEJSA")
		fmt.Println("args:", args)
		execCmd := exec.CommandContext(ctx, "bash", "-c", strings.Join(args, " "))

		finalEnvs := append(secretEnvs, fmt.Sprintf("SECRET_PATH=%s", config.Config.Output))
		finalEnvs = append(os.Environ(), finalEnvs...)
		execCmd.Env = finalEnvs

		if err := util.RunCmdPTY(execCmd, secretEnvs, redact); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Clean up the temporary directory manually before os.Exit since defer won't run
				os.RemoveAll(dir) //nolint:errcheck

				code := exitErr.ExitCode()

				os.Exit(code)
			}
			log.Fatal().Err(err).Msg("Command execution failed")
		}

	},
}

var redact bool

func init() {
	devCmd.PersistentFlags().BoolVar(&redact, "redact", false, "`Redact secrets from output, defaults to false`")

	rootCmd.AddCommand(devCmd)
}
