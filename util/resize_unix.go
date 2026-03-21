//go:build !windows

package util

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
)

func handleResize(ptyFile *os.File) func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			// Will try and inherit the size, we don't really care if it fails.
			pty.InheritSize(os.Stdin, ptyFile) //nolint:errcheck
		}
	}()
	ch <- syscall.SIGWINCH // initial resize trigger

	return func() {
		signal.Stop(ch)
		close(ch)
	}
}
