//go:build windows

package util

import (
	pty "github.com/aymanbagabas/go-pty"
)

func handleResize(_ pty.Pty) func() {
	// Windows does not support SIGWINCH signals in the same way.
	// We return a no-op cleanup function for the Windows build.
	return func() {}
}
