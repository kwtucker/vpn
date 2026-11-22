// Package execx provides the external-command runner used by the VPN CLI.
package execx

import (
	"context"
	"io"
	"os/exec"
)

// Runner executes an external command with the supplied streams. Tests replace
// it with a fake so command order and arguments can be asserted.
type Runner interface {
	Run(ctx context.Context, stdin io.Reader, name string, args []string, stdout, stderr io.Writer) error
}

// OSRunner executes commands with os/exec.
type OSRunner struct{}

// Run implements Runner.
func (OSRunner) Run(ctx context.Context, stdin io.Reader, name string, args []string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
