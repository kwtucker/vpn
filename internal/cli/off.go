package cli

import (
	"context"
	"errors"
	"fmt"
)

// DisconnectAll brings down whichever VPNs are active. Failures from either
// side are reported together so one does not hide the other.
func (a *App) DisconnectAll(ctx context.Context) error {
	var errs []error
	if err := a.ensureTailscaleDown(ctx); err != nil {
		errs = append(errs, err)
	}
	if err := a.ensureMullvadDisconnected(ctx); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	fmt.Fprintln(a.Out, "Both VPNs are disconnected.")
	return nil
}
