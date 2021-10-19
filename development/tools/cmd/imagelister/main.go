package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/kyma-project/test-infra/development/tools/pkg/imagelister"
)

var (
	kymaResourcesDirectory = flag.String("kymaDirectory", "/home/prow/go/src/github.com/kyma-project/kyma/resources/", "Path to Kyma resources")
	outputJSON             = flag.String("o", "", "Path with output JSON file")
)

func main() {
	// map image name to list of components that are using it
	imageComponents := make(imagelister.ImageComponents)

	flag.Parse()
	fmt.Printf("Looking for images in \"%s\"\n\n", *kymaResourcesDirectory)

	var images []imagelister.Image
	var testImages []imagelister.Image

	err := filepath.Walk(*kymaResourcesDirectory, getWalkFunc(&images, &testImages, imageComponents))
	if err != nil {
		fmt.Printf("Cannot traverse directory: %s\n", err)
		os.Exit(2)
	}

	fmt.Println("images:")
	printImages(images, imageComponents)

	fmt.Println("test images:")
	printImages(testImages, imageComponents)

	if *outputJSON != "" {
		err = imagelister.WriteImagesJSON(*outputJSON, images, testImages, imageComponents)
		if err != nil {
			fmt.Printf("Cannot save JSON: %s\n", err)
			os.Exit(2)
		}
	}
}

func printImages(images []imagelister.Image, imageComponents imagelister.ImageComponents) {
	sort.Slice(images, imagelister.GetSortImagesFunc(images))
	for _, image := range images {
		components := imageComponents[image.String()]
		fmt.Printf("%s, used by %s\n", image, strings.Join(components, ", "))
	}
}

func getWalkFunc(images, testImages *[]imagelister.Image, imageComponents imagelister.ImageComponents) filepath.WalkFunc {
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

		var parsedFile imagelister.ValueFile

		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(yamlFile, &parsedFile)
		if err != nil {
			return err
		}

		// TODO get component
		component := strings.Replace(path, *kymaResourcesDirectory, "", -1)
		component = strings.Replace(component, "/values.yaml", "", -1)

		for _, image := range parsedFile.Global.Images {
			// add registry info directly into the image struct
			if image.ContainerRegistryPath == "" {
				image.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
			}
			// remove duplicates
			if !imagelister.ImageListContains(*images, image) {
				*images = append(*images, image)
			}
			imageComponents[image.String()] = append(imageComponents[image.String()], component)
		}

		for _, testImage := range parsedFile.Global.TestImages {
			if testImage.ContainerRegistryPath == "" {
				testImage.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
			}
			if !imagelister.ImageListContains(*testImages, testImage) {
				*testImages = append(*testImages, testImage)
			}
			imageComponents[testImage.String()] = append(imageComponents[testImage.String()], component)
		}

		return nil
	}
}
