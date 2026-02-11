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
			panic(err)
		}
		defer os.RemoveAll(dir) //nolint:errcheck // We don't care if it fails to remove the folder as it is created a a temp folder and the OS should take care of it.

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
		// 4. Start the child application with the temporary file path using the context
		// cmd := exec.CommandContext(ctx, "bash", "-c", "echo $HEJSA")
		fmt.Println("args:", args)
		execCmd := exec.CommandContext(ctx, "bash", "-c", strings.Join(args, " "))

		finalEnvs := append(secretEnvs, fmt.Sprintf("SECRET_PATH=%s", config.Config.Output))
		finalEnvs = append(os.Environ(), finalEnvs...)
		execCmd.Env = finalEnvs

		execCmd.Stdin = os.Stdin
		execCmd.Stdout = &util.Redactor{
			Writer: os.Stdout,
			Envs:   secretEnvs,
			Redact: redact,
		}
		execCmd.Stderr = &util.Redactor{
			Writer: os.Stderr,
			Envs:   secretEnvs,
			Redact: redact,
		}

		if err := execCmd.Run(); err != nil {
			if ctx.Err() != nil || err == context.Canceled {
				// Context cancelled (e.g., ctrl+c)
				return
			}
			panic(err)
		}

	},
}

var redact bool

func init() {
	devCmd.PersistentFlags().BoolVar(&redact, "redact", false, "`Redact secrets from output, defaults to false`")

	rootCmd.AddCommand(devCmd)
}
