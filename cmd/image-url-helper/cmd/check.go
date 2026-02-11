package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-project/test-infra/pkg/image-url-helper/check"
	"github.com/kyma-project/test-infra/pkg/image-url-helper/imagelister"
	imgs "github.com/kyma-project/test-infra/pkg/image-url-helper/images"

	"github.com/jamiealquiza/envy"
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
		//nolint:revive
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

			images := make(imgs.ComponentImageMap)
			testImages := make(imgs.ComponentImageMap)
			err = filepath.Walk(ResourcesDirectory, imagelister.GetWalkFunc(ResourcesDirectoryClean, images, testImages))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			allImages := make(imgs.ComponentImageMap)
			imgs.MergeImageMap(allImages, images)
			imgs.MergeImageMap(allImages, testImages)

			inconsistentImages := imgs.GetInconsistentImages(allImages)

			if len(inconsistentImages) > 0 {
				fmt.Printf("\n--------------------\n")
				fmt.Println("Images with multiple tags:")
				imgs.PrintImages(inconsistentImages)
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
