package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/check"
	"github.com/spf13/cobra"
)

type checkCmdOptions struct {
	skipComments bool
}

func CheckCmd() *cobra.Command {
	options := checkCmdOptions{}
	cmd := &cobra.Command{
		Use:     "check",
		Short:   "Check if all images use new format",
		Long:    "Find all image usages that doesn't use imageurl template",
		Example: "image-url-helper list",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			foundIncompatible := false

			skipComentsInfo := ""
			if options.skipComments {
				skipComentsInfo = ", excluding commented out lines"
			}

			// for all files in resources
			fmt.Printf("Looking for incompatible images in \"%s\"%s:\n\n", ResourcesDirectory, skipComentsInfo)

			err := filepath.Walk(ResourcesDirectory, check.GetkWalkFunc(&foundIncompatible, options.skipComments))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s", err)
				os.Exit(2)
			}

			if foundIncompatible {
				fmt.Printf("\nFound incompatible image lines\n")
				os.Exit(3)
			} else {
				fmt.Println("All images seems to be in the new format")
			}
		},
	}
	addCheckCmdFlags(cmd, &options)
	return cmd
}

func addCheckCmdFlags(cmd *cobra.Command, options *checkCmdOptions) {
	cmd.Flags().BoolVarP(&options.skipComments, "skipComments", "s", false, "Skip commented out lines")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
