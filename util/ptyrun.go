package util

import (
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
func RunCmdPTY(cmd *exec.Cmd, secretEnvs []string, redact bool) error {
	// Start the command with a pseudo-terminal.
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = ptyFile.Close() }() // Best effort cleanup

	// Check if stdin is a terminal. If not, we skip terminal-specific features.
	isTerm := term.IsTerminal(int(os.Stdin.Fd()))

	if isTerm {
		// Handle standard input window resizing
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				// Will try and inherit the size, we don't really care if it fails.
				pty.InheritSize(os.Stdin, ptyFile) //nolint:errcheck // It isn't that important if it fails
			}
		}()
		ch <- syscall.SIGWINCH // initial resize trigger
		defer func() {
			signal.Stop(ch)
		}()

		// Put the true os.Stdin into raw mode to capture Ctrl+C, etc.
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Important: Restore on exit
	}

	// Copy os.Stdin directly to the pseudo-terminal
	// Note: This goroutine intentionally leaks when the command finishes,
	// as standard input reads block indefinitely.
	go func() {
		_, _ = io.Copy(ptyFile, os.Stdin)
	}()

	// The pseudo-terminal unites both stdout and stderr and gives it to us here.
	redactor := &Redactor{
		Writer: os.Stdout,
		Envs:   secretEnvs,
		Redact: redact,
	}

	// Copy the ptyFile output through our redactor back to os.Stdout
	_, err = io.Copy(redactor, ptyFile)
	if err != nil {
		// On some platforms (like Linux), closing the PTY from the child process
		// returns an EIO error. We can safely ignore it.
		pathErr, ok := err.(*os.PathError)
		if !ok || pathErr.Err != syscall.EIO {
			// It's a real error, print it but don't fail the command yet.
			// The caller typically cares more about the exit code.
			os.Stderr.WriteString("error reading pty output: " + err.Error() + "\n") //nolint:errcheck // We are already in an error state, don't care if this also fails
		}
	}

	// Wait for the command to terminate to return its exit error mapping.
	return cmd.Wait()
}
