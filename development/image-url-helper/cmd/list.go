package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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

			err := filepath.Walk(ResourcesDirectory, getListWalkFunc(&images, &testImages, imageComponents))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			fmt.Println("images:")
			printImages(images, imageComponents)

			fmt.Println("test images:")
			printImages(testImages, imageComponents)

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
	cmd.Flags().StringVarP(&options.outputFilename, "outputFilename", "o", "", "Skip commented out lines")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}

func printImages(images []list.Image, imageComponents list.ImageComponents) {
	sort.Slice(images, list.GetSortImagesFunc(images))
	for _, image := range images {
		components := imageComponents[image.String()]
		fmt.Printf("%s, used by %s\n", image, strings.Join(components, ", "))
	}
}

func getListWalkFunc(images, testImages *[]list.Image, imageComponents list.ImageComponents) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		// TODO limit walking to one level?

		//pass the error further, this shouldn't ever happen
		if err != nil {
			return err
		}

		// skip directory entries, we just want files
		if info.IsDir() {
			return nil
		}

		// we only want to check values.yaml files
		if info.Name() != "values.yaml" {
			return nil
		}

		var parsedFile list.ValueFile

		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(yamlFile, &parsedFile)
		if err != nil {
			return err
		}

		// TODO get component
		component := strings.Replace(path, ResourcesDirectory, "", -1)
		component = strings.Replace(component, "/values.yaml", "", -1)

		for _, image := range parsedFile.Global.Images {
			// add registry info directly into the image struct
			if image.ContainerRegistryPath == "" {
				image.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
			}
			// remove duplicates
			if !list.ImageListContains(*images, image) {
				*images = append(*images, image)
			}
			imageComponents[image.String()] = append(imageComponents[image.String()], component)
		}

		for _, testImage := range parsedFile.Global.TestImages {
			if testImage.ContainerRegistryPath == "" {
				testImage.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
			}
			if !list.ImageListContains(*testImages, testImage) {
				*testImages = append(*testImages, testImage)
			}
			imageComponents[testImage.String()] = append(imageComponents[testImage.String()], component)
		}

		return nil
	}
}
