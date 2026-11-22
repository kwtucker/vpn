package cmd

// Execute runs the standalone CLI and returns any execution error.
func Execute(version string) error {
	return NewRootCmd(version).Execute()
}
