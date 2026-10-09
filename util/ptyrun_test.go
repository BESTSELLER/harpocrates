package util

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
)

func TestRunCommand_Success(t *testing.T) {
	ctx := context.Background()
	var name string
	var args []string

	if runtime.GOOS == "windows" {
		name = "cmd.exe"
		args = []string{"/c", "echo hello, world"}
	} else {
		name = "sh"
		args = []string{"-c", "echo 'hello, world'"}
	}

	err := RunCommand(ctx, name, args, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunCommand_ExitCode(t *testing.T) {
	ctx := context.Background()
	var name string
	var args []string

	if runtime.GOOS == "windows" {
		name = "cmd.exe"
		args = []string{"/c", "exit 42"}
	} else {
		name = "sh"
		args = []string{"-c", "exit 42"}
	}

	err := RunCommand(ctx, name, args, nil)
	if err == nil {
		t.Fatalf("expected error due to non-zero exit code, got nil")
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 42 {
			t.Errorf("expected exit code 42, got %d", exitErr.ExitCode())
		}
	} else {
		t.Errorf("expected exec.ExitError, got %T: %v", err, err)
	}
}
