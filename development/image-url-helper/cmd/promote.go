package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/promote"
	"github.com/spf13/cobra"
)

type promoteCmdOptions struct {
	targetContainerRegistry string
	targetTag               string
}

// PromoteCmd replaces containerRegistry and image versions with the provided ones
func PromoteCmd() *cobra.Command {
	options := promoteCmdOptions{}
	cmd := &cobra.Command{
		Use:     "promote",
		Short:   "Promote images",
		Long:    "Replace container registry and image version values in values.yaml files with selected ones",
		Example: "image-url-helper promote --target-container-registry abc --target-tag release-1",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

			if options.targetContainerRegistry == "" && options.targetTag == "" {
				fmt.Println("At leat one flag expected, nothing to do")
				cmd.Help()
				os.Exit(1)
			}

			err := filepath.Walk(ResourcesDirectory, promote.GetWalkFunc(ResourcesDirectoryClean, options.targetContainerRegistry, options.targetTag))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

		},
	}
	addPromoteCmdFlags(cmd, &options)
	return cmd
}

func addPromoteCmdFlags(cmd *cobra.Command, options *promoteCmdOptions) {
	cmd.Flags().StringVarP(&options.targetContainerRegistry, "target-container-registry", "c", "", "Name of the target registry")
	cmd.Flags().StringVarP(&options.targetTag, "target-tag", "t", "", "Name of the target tag")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
