package cli

import "github.com/spf13/cobra"

// NewRootCmd builds the complete command tree. The public cmd package wraps it
// for the standalone binary and for embedding in other Cobra CLIs.
func NewRootCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:   "vpn",
		Short: "Switch between Mullvad and Tailscale without running both",
		Long: `Connect to exactly one VPN at a time.

'vpn mull' brings Tailscale down before connecting Mullvad.
'vpn tail' disconnects Mullvad before bringing Tailscale up.
Extra flags after 'mull' or 'tail' go straight to the underlying CLI.`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.Status(cmd.Context())
		},
	}
	if app.Version != "" {
		root.Version = app.Version
		root.SetVersionTemplate("vpn {{.Version}}\n")
	}
	root.SetIn(app.In)
	root.SetOut(app.Out)
	root.SetErr(app.Err)
	root.AddCommand(
		newMullCmd(app),
		newTailCmd(app),
		newOffCmd(app),
		newStatusCmd(app),
	)
	return root
}

func newMullCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "mull [mullvad connect flags]",
		Aliases: []string{"m", "mullvad"},
		Short:   "Connect Mullvad (Tailscale is brought down first)",
		// Extra flags are forwarded verbatim; handle --help ourselves.
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if wantsHelp(args) {
				return cmd.Help()
			}
			return app.ConnectMullvad(cmd.Context(), args)
		},
	}
}

func newTailCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:                "tail [tailscale up flags]",
		Aliases:            []string{"t", "tailscale"},
		Short:              "Bring Tailscale up (Mullvad is disconnected first)",
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if wantsHelp(args) {
				return cmd.Help()
			}
			return app.ConnectTailscale(cmd.Context(), args)
		},
	}
}

func newOffCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "off",
		Aliases: []string{"down", "disconnect"},
		Short:   "Disconnect both VPNs",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.DisconnectAll(cmd.Context())
		},
	}
}

func newStatusCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "Show both VPN states (also the default with no arguments)",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.Status(cmd.Context())
		},
	}
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}
