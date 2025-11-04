package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// SetupCommandSilence configures a cobra.Command to suppress automatic usage and error printing, returning a formatted usage error explicitly.
func SetupCommandSilence(rootCmd *cobra.Command) {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("wrong command usage: %w\n%s", err, c.UsageString())
	})
}
