package cmd

import (
	"fmt"
	"github.com/kyma-project/test-infra/pkg/image-url-helper/common"
	"github.com/kyma-project/test-infra/pkg/image-url-helper/list"
	"os"
	"path/filepath"

	"github.com/jamiealquiza/envy"
	"github.com/spf13/cobra"
)

type listCmdOptions struct {
	outputFormat      string
	excludeTestImages bool
}

// ListCmd lists all images defined in values.yaml files
func ListCmd() *cobra.Command {
	options := listCmdOptions{}
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all images used in charts",
		Long:    "List all images used in Helm charts by checking values.yaml files",
		Example: "image-url-helper list",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

			images := make(common.ComponentImageMap)
			testImages := make(common.ComponentImageMap)

			err := filepath.Walk(ResourcesDirectory, list.GetWalkFunc(ResourcesDirectoryClean, images, testImages))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			allImages := make(common.ComponentImageMap)
			common.MergeImageMap(allImages, images)
			if !options.excludeTestImages {
				common.MergeImageMap(allImages, testImages)
			}

			err = common.PrintComponentImageMap(allImages, options.outputFormat)
			if err != nil {
				common.PrintAndFail(3, "Cannot print image list: %s\n", err)
			}
		},
	}
	addListCmdFlags(cmd, &options)
	return cmd
}

func addListCmdFlags(cmd *cobra.Command, options *listCmdOptions) {
	cmd.Flags().StringVarP(&options.outputFormat, "output-format", "o", "", "Name of the output format (json/yaml)")
	cmd.Flags().BoolVarP(&options.excludeTestImages, "exclude-test-images", "e", false, "Exclude test images from the output list")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
