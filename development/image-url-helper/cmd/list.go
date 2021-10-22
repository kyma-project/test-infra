package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"github.com/spf13/cobra"
)

type listCmdOptions struct {
	outputFilename string
}

func ListCmd() *cobra.Command {
	options := listCmdOptions{}
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all images used in charts",
		Long:    "List all images used in Helm charts by checking values.yaml files",
		Example: "image-url-helper list",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			imageComponents := make(list.ImageComponents)

			fmt.Printf("Looking for images in \"%s\"\n\n", ResourcesDirectory)

			var images []list.Image
			var testImages []list.Image

			err := filepath.Walk(ResourcesDirectory, list.GetWalkFunc(ResourcesDirectory, &images, &testImages, imageComponents))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			var allImages []list.Image
			allImages = append(allImages, images...)
			allImages = append(allImages, testImages...)
			sort.Slice(allImages, list.GetSortImagesFunc(allImages))
			allImages = list.RemoveDoubles(allImages)

			fmt.Println("images:")
			list.PrintImages(allImages, imageComponents)

			inconsistentImages := list.GetInconsistentImages(allImages)

			fmt.Println("Inconsistent images:")
			list.PrintImages(inconsistentImages, imageComponents)

			if options.outputFilename != "" {
				err = list.WriteImagesJSON(options.outputFilename, images, testImages, imageComponents)
				if err != nil {
					fmt.Printf("Cannot save JSON: %s\n", err)
					os.Exit(2)
				}
			}
		},
	}
	addListCmdFlags(cmd, &options)
	return cmd
}

func addListCmdFlags(cmd *cobra.Command, options *listCmdOptions) {
	cmd.Flags().StringVarP(&options.outputFilename, "outputFilename", "o", "", "Name of the output JSON file")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
