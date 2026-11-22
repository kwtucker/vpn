package cli

import (
	"context"
	"errors"
	"fmt"
)

// Status prints both states. If both are active it reports a conflict, so the
// wrapper never silently accepts a state it is designed to prevent.
func (a *App) Status(ctx context.Context) error {
	mullvad, err := a.mullvadConnected(ctx)
	if err != nil {
		return err
	}
	tailscale, err := a.tailscaleRunning(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.Out, "Mullvad:   %s\n", stateLabel(mullvad, "connected", "disconnected"))
	fmt.Fprintf(a.Out, "Tailscale: %s\n", stateLabel(tailscale, "up", "down"))
	if mullvad && tailscale {
		fmt.Fprintln(a.Err, "CONFLICT: both are active; run 'vpn off' to disconnect both")
		return errors.New("both Mullvad and Tailscale are active")
	}
	return nil
}

func stateLabel(active bool, activeText, inactiveText string) string {
	if active {
		return activeText
	}
	return inactiveText
}
