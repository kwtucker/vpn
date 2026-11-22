// Package cmd exposes the vpn command tree for the standalone binary and for
// embedding in other Cobra CLIs such as wb.
package cmd

import (
	"os"

	"github.com/kwtucker/vpn/internal/cli"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the vpn command tree with OS-backed execution. An empty
// version suppresses the --version flag (useful when embedded in wb).
func NewRootCmd(version string) *cobra.Command {
	app := cli.NewApp(nil, os.Stdin, os.Stdout, os.Stderr)
	app.Version = version
	return cli.NewRootCmd(app)
}
