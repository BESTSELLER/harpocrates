package util

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// RunCmdPTY runs the given command in a pseudo-terminal to preserve colored output,
// wiring its output through the Redactor to ensure sensitive values are stripped.
func RunCmdPTY(ctx context.Context, cmd *exec.Cmd, secretEnvs []string, redact bool) error {
	// Start the command with a pseudo-terminal.
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = ptmx.Close() }() // Best effort cleanup

	// Handle standard input window resizing
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				// failed to inherit size, usually fine (ignorable)
			}
		}
	}()
	ch <- syscall.SIGWINCH // initial resize trigger
	defer func() {
		signal.Stop(ch)
		close(ch)
	}()

	// Put the true os.Stdin into raw mode to capture Ctrl+C, etc.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Important: Restore on exit

	// Copy os.Stdin directly to the pseudo-terminal
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	// The pseudo-terminal unites both stdout and stderr and gives it to us here.
	redactor := &Redactor{
		Writer: os.Stdout,
		Envs:   secretEnvs,
		Redact: redact,
	}

	// Copy the ptmx output through our redactor back to os.Stdout
	_, _ = io.Copy(redactor, ptmx)

	// Wait for the command to terminate to return its exit error mapping.
	return cmd.Wait()
}
