// Package cli implements the vpn command tree and its VPN operations.
package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kwtucker/vpn/internal/execx"
)

// App holds dependencies. Tests supply a fake Runner and buffers.
type App struct {
	Runner  execx.Runner
	Version string
	In      io.Reader
	Out     io.Writer
	Err     io.Writer

	// Timeouts and poll interval. Tests shorten these.
	CommandTimeout time.Duration
	ConnectTimeout time.Duration
	DownTimeout    time.Duration
	WaitInterval   time.Duration
}

// NewApp builds an App with sensible defaults and OS-backed execution.
func NewApp(runner execx.Runner, in io.Reader, out, errOut io.Writer) *App {
	if runner == nil {
		runner = execx.OSRunner{}
	}
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	return &App{
		Runner:         runner,
		In:             in,
		Out:            out,
		Err:            errOut,
		CommandTimeout: 60 * time.Second,
		ConnectTimeout: 30 * time.Second,
		DownTimeout:    10 * time.Second,
		WaitInterval:   100 * time.Millisecond,
	}
}

func (a *App) run(ctx context.Context, name string, args ...string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, a.CommandTimeout)
	defer cancel()
	return a.Runner.Run(cmdCtx, a.In, name, args, a.Out, a.Err)
}

// output captures a status command's JSON. Child stderr is kept for errors but
// never printed on success, so status checks stay silent.
func (a *App) output(ctx context.Context, name string, args ...string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, a.CommandTimeout)
	defer cancel()
	var stdout, stderr bytes.Buffer
	if err := a.Runner.Run(cmdCtx, nil, name, args, &stdout, &stderr); err != nil {
		return "", commandError(err, &stdout, &stderr)
	}
	return stdout.String(), nil
}

// runQuiet runs a command without streaming its progress. On failure the
// captured output is included in the error so nothing is lost.
func (a *App) runQuiet(ctx context.Context, name string, args ...string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, a.CommandTimeout)
	defer cancel()
	var stdout, stderr bytes.Buffer
	if err := a.Runner.Run(cmdCtx, a.In, name, args, &stdout, &stderr); err != nil {
		return commandError(err, &stdout, &stderr)
	}
	return nil
}

func commandError(err error, stdout, stderr *bytes.Buffer) error {
	detail := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	if detail == "" {
		return err
	}
	return fmt.Errorf("%w: %s", err, detail)
}
