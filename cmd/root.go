// Package cmd exposes the vpn command tree for the standalone binary and for
// embedding in other Cobra CLIs.
package cmd

import (
	"os"

	"github.com/kwtucker/vpn/internal/cli"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the vpn command tree with OS-backed execution. An empty
// version suppresses the --version flag when embedded in another CLI.
func NewRootCmd(version string) *cobra.Command {
	app := cli.NewApp(nil, os.Stdin, os.Stdout, os.Stderr)
	app.Version = version
	return cli.NewRootCmd(app)
}
