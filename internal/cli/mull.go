package cli

import (
	"context"
	"fmt"
)

// ConnectMullvad brings Tailscale down first, then connects Mullvad and
// verifies that exactly one VPN is active when it returns.
func (a *App) ConnectMullvad(ctx context.Context, extra []string) error {
	if err := a.ensureTailscaleDown(ctx); err != nil {
		return err
	}

	args := []string{"connect"}
	if !hasFlag(extra, "-w", "--wait") {
		args = append(args, "--wait")
	}
	args = append(args, extra...)
	if err := a.run(ctx, "mullvad", args...); err != nil {
		return fmt.Errorf("mullvad connect: %w", err)
	}
	if err := a.waitForMullvad(ctx, true, a.ConnectTimeout); err != nil {
		_ = a.run(ctx, "mullvad", "disconnect", "--wait")
		return fmt.Errorf("mullvad did not connect: %w", err)
	}
	if err := a.ensureTailscaleDown(ctx); err != nil {
		return err
	}
	fmt.Fprintln(a.Out, "Mullvad connected; Tailscale is down.")
	return nil
}

func (a *App) ensureTailscaleDown(ctx context.Context) error {
	running, err := a.tailscaleRunning(ctx)
	if err != nil {
		return err
	}
	if !running {
		return nil
	}
	if err := a.run(ctx, "tailscale", "down"); err != nil {
		return fmt.Errorf("tailscale down: %w", err)
	}
	if err := a.waitForTailscale(ctx, false, a.DownTimeout); err != nil {
		return fmt.Errorf("tailscale did not go down: %w", err)
	}
	return nil
}
