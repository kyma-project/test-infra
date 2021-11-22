package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/check"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"github.com/spf13/cobra"
)

type checkCmdOptions struct {
	skipComments bool
	excludesList string
}

// CheckCmd checks image definitions and images with multiple tags
func CheckCmd() *cobra.Command {
	options := checkCmdOptions{}
	cmd := &cobra.Command{
		Use:     "check",
		Short:   "Check if all images use new format",
		Long:    "Find all image usages that doesn't use imageurl template",
		Example: "image-url-helper list",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

			var imagesDefinedOutside []check.ImageLine

			excludes, err := check.ParseExcludes(options.excludesList)
			if err != nil {
				fmt.Printf("Cannot parse excludes list: %s\n", err)
				os.Exit(2)
			}

			err = filepath.Walk(ResourcesDirectoryClean, check.GetkWalkFunc(ResourcesDirectoryClean, &imagesDefinedOutside, options.skipComments, excludes))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			if len(imagesDefinedOutside) > 0 {
				fmt.Println("Images defined outside of values.yaml:")
				for _, image := range imagesDefinedOutside {
					fmt.Printf("%s:%d: %s\n", image.Filename, image.LineNumber, image.Line)
				}
			}

			var images []list.Image
			var testImages []list.Image
			imageComponentsMap := make(list.ImageToComponents)
			err = filepath.Walk(ResourcesDirectory, list.GetWalkFunc(ResourcesDirectoryClean, &images, &testImages, imageComponentsMap))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			var allImages []list.Image
			allImages = append(allImages, images...)
			allImages = append(allImages, testImages...)
			sort.Slice(allImages, list.GetSortImagesFunc(allImages))
			allImages = list.RemoveDoubles(allImages)

			inconsistentImages := list.GetInconsistentImages(allImages)

			if len(inconsistentImages) > 0 {
				fmt.Printf("\n--------------------\n")
				fmt.Println("Images with multiple tags:")
				list.PrintImages(inconsistentImages, imageComponentsMap)
			}
			if len(imagesDefinedOutside) > 0 || len(inconsistentImages) > 0 {
				os.Exit(3)
			}
		},
	}
	addCheckCmdFlags(cmd, &options)
	return cmd
}

func addCheckCmdFlags(cmd *cobra.Command, options *checkCmdOptions) {
	cmd.Flags().BoolVarP(&options.skipComments, "skip-comments", "s", true, "Skip commented out lines")
	cmd.Flags().StringVarP(&options.excludesList, "excludes-list", "e", "", "List of excluded images")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}
