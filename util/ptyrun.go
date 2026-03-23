package util

import (
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// RunCmdPTY runs the given command in a pseudo-terminal
func RunCmdPTY(cmd *exec.Cmd, secretEnvs []string, redact bool) error {
	// Start the command with a pseudo-terminal.
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = ptyFile.Close() }() // Best effort cleanup - we don't care about errors here.

	// Check if stdin is a terminal. If not, we skip terminal-specific features.
	isTerm := term.IsTerminal(int(os.Stdin.Fd()))

	if isTerm {
		// Handle standard input window resizing
		cleanupResize := handleResize(ptyFile)
		defer cleanupResize()

		// Put the true os.Stdin into raw mode to capture Ctrl+C, etc.
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			// Will try cleanup up the process so that we don't have any zombie processes.
			cmd.Process.Kill() //nolint:errcheck // We are already in an error state, don't care if this also fails
			cmd.Wait()         //nolint:errcheck // We are already in an error state, don't care if this also fails
			return err
		}
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Important: Restore on exit
	}

	// Copy os.Stdin directly to the pseudo-terminal
	go func() {
		_, _ = io.Copy(ptyFile, os.Stdin)
	}()

	redactor := &Redactor{
		Writer: os.Stdout,
		Envs:   secretEnvs,
		Redact: redact,
	}

	// Copy the ptyFile output through our redactor back to os.Stdout
	_, err = io.Copy(redactor, ptyFile)
	if err != nil {
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
