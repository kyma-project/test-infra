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

// ListCmd lists all images defined in values.yaml files
func PromoteCmd() *cobra.Command {
	options := promoteCmdOptions{}
	cmd := &cobra.Command{
		Use: "promote",
		// TODO this description is horrible
		Short:   "Promote images",
		Long:    "List all images used in Helm charts by checking values.yaml files",
		Example: "image-url-helper promote --target-container-registry abc",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

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
	// TODO add possiblity to do not-inplace replacements???
	cmd.Flags().StringVarP(&options.targetContainerRegistry, "target-container-registry", "c", "", "Name of the target registry")
	cmd.Flags().StringVarP(&options.targetTag, "target-tag", "t", "", "Name of the target tag")
	cmd.MarkFlagRequired("target-container-registry")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
