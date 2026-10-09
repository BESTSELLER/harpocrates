package util

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"syscall"

	pty "github.com/aymanbagabas/go-pty"
	"golang.org/x/term"
)

// RunCommand runs the given command. If stdin is a terminal, it runs inside a pseudo-terminal (PTY/ConPTY);
// otherwise, it falls back to standard process execution with standard I/O pipes.
func RunCommand(ctx context.Context, name string, args []string, env []string) error {
	isTerm := term.IsTerminal(int(os.Stdin.Fd()))
	if !isTerm {
		return runPipe(ctx, name, args, env)
	}

	p, err := pty.New()
	if err != nil {
		// Failed to allocate pseudo-terminal (e.g. ConPTY unavailable on older Windows), fall back to pipe execution
		return runPipe(ctx, name, args, env)
	}

	if w, h, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
		_ = p.Resize(w, h)
	}

	cleanupResize := handleResize(p)
	defer cleanupResize()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		_ = p.Close()
		return runPipe(ctx, name, args, env)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	cmd := p.CommandContext(ctx, name, args...)
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		_ = p.Close()
		return err
	}

	go func() {
		_, _ = io.Copy(p, os.Stdin)
	}()

	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		_, err := io.Copy(os.Stdout, p)
		if err != nil {
			var pathErr *os.PathError
			if !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) && !errors.Is(err, syscall.EIO) && (!errors.As(err, &pathErr) || (pathErr.Err != syscall.EIO && pathErr.Err != os.ErrClosed)) {
				os.Stderr.WriteString("error reading pty output: " + err.Error() + "\n") //nolint:errcheck
			}
		}
	}()

	waitErr := cmd.Wait()

	_ = p.Close()
	<-outputDone

	return waitErr
}

func runPipe(ctx context.Context, name string, args []string, env []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
