package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type mullvadStatus struct {
	State string `json:"state"`
}

type tailscaleStatus struct {
	BackendState string `json:"BackendState"`
}

// mullvadConnected reports whether the Mullvad tunnel is up. Any state other
// than "disconnected" (for example "connecting") counts as up, so a switch is
// never made while a transition is still in progress.
func (a *App) mullvadConnected(ctx context.Context) (bool, error) {
	out, err := a.output(ctx, "mullvad", "status", "--json")
	if err != nil {
		return false, fmt.Errorf("mullvad status: %w", err)
	}
	var status mullvadStatus
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		return false, fmt.Errorf("parse mullvad status: %w", err)
	}
	return status.State != "disconnected", nil
}

// tailscaleRunning reports whether the tailscale backend is Running.
func (a *App) tailscaleRunning(ctx context.Context) (bool, error) {
	out, err := a.output(ctx, "tailscale", "status", "--json")
	if err != nil {
		return false, fmt.Errorf("tailscale status: %w", err)
	}
	var status tailscaleStatus
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		return false, fmt.Errorf("parse tailscale status: %w", err)
	}
	return status.BackendState == "Running", nil
}

func (a *App) waitForMullvad(ctx context.Context, want bool, timeout time.Duration) error {
	return a.wait(ctx, timeout, "Mullvad", func() (bool, error) {
		return a.mullvadConnected(ctx)
	}, want)
}

func (a *App) waitForTailscale(ctx context.Context, want bool, timeout time.Duration) error {
	return a.wait(ctx, timeout, "Tailscale", func() (bool, error) {
		return a.tailscaleRunning(ctx)
	}, want)
}

func (a *App) wait(ctx context.Context, timeout time.Duration, name string, check func() (bool, error), want bool) error {
	deadline := time.Now().Add(timeout)
	for {
		got, err := check()
		if err != nil {
			return err
		}
		if got == want {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for %s: got %v, want %v", name, got, want)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(a.WaitInterval):
		}
	}
}

// hasFlag reports whether args contain the given flag, allowing either
// "--flag" or "--flag=value" forms.
func hasFlag(args []string, names ...string) bool {
	for _, arg := range args {
		for _, name := range names {
			if arg == name || strings.HasPrefix(arg, name+"=") {
				return true
			}
		}
	}
	return false
}
