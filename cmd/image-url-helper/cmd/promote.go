package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-project/test-infra/pkg/image-url-helper/common"
	"github.com/kyma-project/test-infra/pkg/image-url-helper/promote"

	"github.com/jamiealquiza/envy"
	"github.com/spf13/cobra"
)

type promoteCmdOptions struct {
	targetContainerRegistry string
	targetTag               string
	dryRun                  bool
	excludesList            string
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
		//nolint:revive
		Run: func(cmd *cobra.Command, args []string) {
			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

			targetContainerRegistryClean := filepath.Clean(options.targetContainerRegistry)

			images := make(common.ComponentImageMap)
			testImages := make(common.ComponentImageMap)

			excludes, err := promote.ParseExcludes(options.excludesList)

			if err != nil {
				fmt.Printf("Cannot parse excludes list: %s\n", err)
				os.Exit(2)
			}

			err = filepath.Walk(ResourcesDirectory, promote.GetWalkFunc(ResourcesDirectoryClean, targetContainerRegistryClean, options.targetTag, options.dryRun, images, testImages, excludes))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			// join both images lists
			allImages := make(common.ComponentImageMap)
			common.MergeImageMap(allImages, images)
			common.MergeImageMap(allImages, testImages)

			err = promote.PrintExternalSyncerYaml(allImages, options.targetTag)
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
	cmd.Flags().StringVarP(&options.excludesList, "excludes-list", "e", "", "Path to the file containing a list of excluded images")
	cmd.MarkFlagRequired("target-container-registry")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
