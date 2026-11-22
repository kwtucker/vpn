package main

import (
	"fmt"
	"os"

	"github.com/kwtucker/vpn/cmd"
)

// version is set by GoReleaser with -X main.version=...
var version = "dev"

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprintf(os.Stderr, "vpn: %v\n", err)
		os.Exit(1)
	}
}
