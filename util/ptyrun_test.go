package util

import (
	"os/exec"
	"testing"
)

func TestRunCmdPTY_Success(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo 'hello, world'")
	secretEnvs := []string{"SUPER_SECRET_ENV"}
	redact := false

	// This should run smoothly and not throw EIO or panics
	// even when standard input is not a real terminal (like in `go test` and CI).
	err := RunCmdPTY(cmd, secretEnvs, redact)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunCmdPTY_ExitCode(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 42")
	secretEnvs := []string{}
	redact := false

	err := RunCmdPTY(cmd, secretEnvs, redact)
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
