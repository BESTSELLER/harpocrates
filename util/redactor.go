package util

import (
	"io"
	"strings"
)

// Redactor wraps an io.Writer to redact sensitive values from the output.
type Redactor struct {
	Writer io.Writer
	Envs   []string
	Redact bool
}

// Write implements the io.Writer interface, redacting sensitive values from p before writing to the underlying Writer.
func (r *Redactor) Write(p []byte) (n int, err error) {
	result := p

	if r.Redact {
		output := string(p)
		for _, env := range r.Envs {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				value := parts[1]
				if value != "" {
					output = strings.ReplaceAll(output, value, "[REDACTED]")
				}
			}
		}
		result = []byte(output)
	}

	_, err = r.Writer.Write(result)
	return len(p), err
}
