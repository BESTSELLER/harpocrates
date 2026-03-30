//go:build windows

package util

import (
	"os"
)

func handleResize(ptyFile *os.File) func() {
	// Windows does not support SIGWINCH signals in the same way.
	// We return a no-op cleanup function for the Windows build.
	return func() {}
}
