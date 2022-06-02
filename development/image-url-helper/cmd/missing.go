package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/common"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/missing"
	"github.com/spf13/cobra"
)

type existsCmdOptions struct {
	outputFormat      string
	excludeTestImages bool
}

func ExistsCmd() *cobra.Command {

	options := existsCmdOptions{}
	cmd := &cobra.Command{
		Use:     "missing",
		Short:   "Check if all images exists",
		Long:    "Find all images taht don't exist",
		Example: "image-url-helper missing",
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

			missingImages := make(common.ComponentImageMap)

			err = missing.CheckForMissingImages(allImages, missingImages)
			if err != nil {
				fmt.Printf("Cannot check for missing images: %s\n", err)
				os.Exit(4)
			}

			if options.outputFormat == "" {
				common.PrintImages(missingImages)
			} else if strings.ToLower(options.outputFormat) == "json" {
				err = list.PrintImagesJSON(missingImages)
				if err != nil {
					fmt.Printf("Cannot save JSON: %s\n", err)
					os.Exit(2)
				}
			} else if strings.ToLower(options.outputFormat) == "yaml" {
				err = list.PrintImagesYAML(missingImages)
				if err != nil {
					fmt.Printf("Cannot save JSON: %s\n", err)
					os.Exit(2)
				}
			} else {
				fmt.Printf("Unknown output format: %s\n", options.outputFormat)
				os.Exit(2)
			}

			if len(missingImages) > 0 {
				os.Exit(3)
			}
		},
	}
	addExistsCmdFlags(cmd, &options)
	return cmd
}

func addExistsCmdFlags(cmd *cobra.Command, options *existsCmdOptions) {
	cmd.Flags().StringVarP(&options.outputFormat, "output-format", "o", "", "Name of the output format (json/yaml)")
	cmd.Flags().BoolVarP(&options.excludeTestImages, "exclude-test-images", "e", false, "Exclude test images from the output list")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
