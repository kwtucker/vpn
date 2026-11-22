// Package cli implements the vpn command tree and its VPN operations.
package cli

import (
	"bytes"
	"context"
	"io"
	"os"
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

func (a *App) output(ctx context.Context, name string, args ...string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, a.CommandTimeout)
	defer cancel()
	var buf bytes.Buffer
	if err := a.Runner.Run(cmdCtx, nil, name, args, &buf, a.Err); err != nil {
		return "", err
	}
	return buf.String(), nil
}
