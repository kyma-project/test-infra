package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/promote"
	"github.com/spf13/cobra"
)

type promoteCmdOptions struct {
	targetContainerRegistry string
	targetTag               string
	dryRun                  bool
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

			var images []list.Image

			err := filepath.Walk(ResourcesDirectory, promote.GetWalkFunc(ResourcesDirectoryClean, options.targetContainerRegistry, options.targetTag, options.dryRun, &images))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			sort.Slice(images, list.GetSortImagesFunc(images))
			images = list.RemoveDoubles(images)
			// TODO pass "sign" as cli arg
			err = promote.PrintExternalSyncerYaml(images, options.targetContainerRegistry, options.targetTag, true)
			if err != nil {
				fmt.Printf("Cannot print list of images: %s\n", err)
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
	cmd.Flags().BoolVarP(&options.dryRun, "dry-run", "d", true, "Dry run enabled, nothing is changed")
	cmd.MarkFlagRequired("target-container-registry")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
