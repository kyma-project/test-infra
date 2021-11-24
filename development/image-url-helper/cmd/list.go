package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
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
			imageComponentsMap := make(list.ImageToComponents)

			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

			images := make(list.ImageMap)
			testImages := make(list.ImageMap)

			err := filepath.Walk(ResourcesDirectory, list.GetWalkFunc(ResourcesDirectoryClean, images, testImages, imageComponentsMap))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			allImages := make(list.ImageMap)
			list.MergeImageMap(allImages, images)
			list.MergeImageMap(allImages, testImages)

			if options.outputFormat == "" {
				list.PrintImages(allImages, imageComponentsMap)
			} else if strings.ToLower(options.outputFormat) == "json" {
				err = list.PrintImagesJSON(allImages, imageComponentsMap)
				if err != nil {
					fmt.Printf("Cannot save JSON: %s\n", err)
					os.Exit(2)
				}
			} else if strings.ToLower(options.outputFormat) == "yaml" {
				err = list.PrintImagesYAML(allImages, imageComponentsMap)
				if err != nil {
					fmt.Printf("Cannot save JSON: %s\n", err)
					os.Exit(2)
				}
			} else {
				fmt.Printf("Unknown output format: %s\n", options.outputFormat)
				os.Exit(2)
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
