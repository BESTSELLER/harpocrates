//go:build !windows

package util

import (
	"os"
	"os/signal"
	"syscall"

	pty "github.com/aymanbagabas/go-pty"
	"golang.org/x/term"
)

func handleResize(p pty.Pty) func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if w, h, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
				_ = p.Resize(w, h)
			}
		}
	}()
	ch <- syscall.SIGWINCH // initial resize trigger

	return func() {
		signal.Stop(ch)
		close(ch)
	}
}
