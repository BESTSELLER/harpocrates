package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
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
		if len(args) == 0 {
			log.Fatal().Msg("No command provided to execute")
		}
		// Create temporary directory and file
		dir, err := os.MkdirTemp("", "harpocrates")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create temporary directory")
		}
		cleanup := func() {
			os.RemoveAll(dir) //nolint:errcheck
		}
		defer cleanup() // Best-effort cleanup of the temp folder; ignore errors as failure is non-critical and the OS will eventually reclaim the space.

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
			cleanup()
			cancel()
		}()

		fmt.Println(config.Config.Output)
		// Start the child application with the temporary file path using the context
		var execCmd *exec.Cmd
		if len(args) == 1 {
			execCmd = exec.CommandContext(ctx, args[0])
		} else {
			execCmd = exec.CommandContext(ctx, args[0], args[1:]...)
		}

		finalEnvs := append(secretEnvs, fmt.Sprintf("SECRET_PATH=%s", config.Config.Output))
		finalEnvs = append(os.Environ(), finalEnvs...)
		execCmd.Env = finalEnvs

		if err := util.RunCmdPTY(execCmd, secretEnvs, redact); err != nil {
			cleanup() // Clean up the temporary directory manually before os.Exit or log.Fatal since defer won't run
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
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
