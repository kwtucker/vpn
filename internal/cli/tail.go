package cli

import (
	"context"
	"fmt"
)

// ConnectTailscale disconnects Mullvad first, then brings Tailscale up and
// verifies that exactly one VPN is active when it returns.
func (a *App) ConnectTailscale(ctx context.Context, extra []string) error {
	if err := a.ensureMullvadDisconnected(ctx); err != nil {
		return err
	}

	// Call plain `tailscale up` unless the user supplies flags. Tailscale treats
	// any flag as a settings change and then requires every non-default pref to
	// be repeated; no flags means "bring the interface up with current prefs".
	args := append([]string{"up"}, extra...)
	if err := a.run(ctx, "tailscale", args...); err != nil {
		return fmt.Errorf("tailscale up: %w", err)
	}
	if err := a.waitForTailscale(ctx, true, a.ConnectTimeout); err != nil {
		_ = a.run(ctx, "tailscale", "down")
		return fmt.Errorf("tailscale did not come up: %w", err)
	}
	if err := a.ensureMullvadDisconnected(ctx); err != nil {
		return err
	}
	fmt.Fprintln(a.Out, "Tailscale up; Mullvad is disconnected.")
	return nil
}

func (a *App) ensureMullvadDisconnected(ctx context.Context) error {
	connected, err := a.mullvadConnected(ctx)
	if err != nil {
		return err
	}
	if !connected {
		return nil
	}
	if err := a.run(ctx, "mullvad", "disconnect", "--wait"); err != nil {
		return fmt.Errorf("mullvad disconnect: %w", err)
	}
	if err := a.waitForMullvad(ctx, false, a.DownTimeout); err != nil {
		return fmt.Errorf("mullvad did not disconnect: %w", err)
	}
	return nil
}
