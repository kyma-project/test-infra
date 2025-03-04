package cmd

import (
	"os"
	"path/filepath"

	"github.com/kyma-project/test-infra/pkg/image-url-helper/common"
	"github.com/kyma-project/test-infra/pkg/image-url-helper/list"
	"github.com/kyma-project/test-infra/pkg/image-url-helper/missing"

	"github.com/jamiealquiza/envy"
	"github.com/spf13/cobra"
)

type missingCmdOptions struct {
	outputFormat      string
	excludeTestImages bool
}

// MissingCmd checks if all images exists
func MissingCmd() *cobra.Command {

	options := missingCmdOptions{}
	cmd := &cobra.Command{
		Use:     "missing",
		Short:   "Check if all images exists",
		Long:    "Find all images that don't exist",
		Example: "image-url-helper missing",
		Args:    cobra.ExactArgs(0),
		//nolint:revive
		Run: func(cmd *cobra.Command, args []string) {
			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

			images := make(common.ComponentImageMap)
			testImages := make(common.ComponentImageMap)

			err := filepath.Walk(ResourcesDirectory, list.GetWalkFunc(ResourcesDirectoryClean, images, testImages))
			if err != nil {
				common.PrintAndFail(1, "Cannot traverse directory: %s\n", err)
			}

			allImages := make(common.ComponentImageMap)
			common.MergeImageMap(allImages, images)
			if !options.excludeTestImages {
				common.MergeImageMap(allImages, testImages)
			}

			missingImages := make(common.ComponentImageMap)

			err = missing.CheckForMissingImages(allImages, missingImages)
			if err != nil {
				common.PrintAndFail(2, "Cannot check for missing images: %s\n", err)
			}

			err = common.PrintComponentImageMap(missingImages, options.outputFormat)
			if err != nil {
				common.PrintAndFail(3, "Cannot print image list: %s\n", err)
			}

			if len(missingImages) > 0 {
				os.Exit(3)
			}
		},
	}
	addExistsCmdFlags(cmd, &options)
	return cmd
}

func addExistsCmdFlags(cmd *cobra.Command, options *missingCmdOptions) {
	cmd.Flags().StringVarP(&options.outputFormat, "output-format", "o", "", "Name of the output format (json/yaml)")
	cmd.Flags().BoolVarP(&options.excludeTestImages, "exclude-test-images", "e", false, "Exclude test images from the output list")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
# (2025-03-04)