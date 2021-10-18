package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v2"

	"github.com/kyma-project/test-infra/development/tools/pkg/imagelister"
)

var (
	kymaResourcesDirectory = flag.String("kymaDirectory", "/home/prow/go/src/github.com/kyma-project/kyma/resources/", "Path to Kyma resources")
)

func main() {
	flag.Parse()
	fmt.Printf("Looking for images in \"%s\"\n\n", *kymaResourcesDirectory)

	var images []imagelister.Image
	var testImages []imagelister.Image

	err := filepath.Walk(*kymaResourcesDirectory, getWalkFunc(&images, &testImages))
	if err != nil {
		fmt.Printf("Cannot traverse directory: %s\n", err)
		os.Exit(2)
	}

	sort.Slice(images, imagelister.GetSortImagesFunc(images))

	fmt.Println("images:")
	// TODO function here
	for _, image := range images {
		fmt.Printf("%s\n", image)
	}

	sort.Slice(testImages, imagelister.GetSortImagesFunc(testImages))

	fmt.Println("test images:")
	// TODO function here
	for _, testImage := range testImages {
		fmt.Printf("%s\n", testImage)
	}
}

func getWalkFunc(images, testImages *[]imagelister.Image) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		// TODO how to limit walking to one level?

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

		for _, image := range parsedFile.Global.Images {
			// add registry info directly into the image struct
			if image.ContainerRegistryPath == "" {
				image.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
			}
			// remove duplicates
			if !imagelister.ImageListContains(*images, image) {
				*images = append(*images, image)
			}
		}

		for _, testImage := range parsedFile.Global.TestImages {
			if testImage.ContainerRegistryPath == "" {
				testImage.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
			}
			if !imagelister.ImageListContains(*testImages, testImage) {
				*testImages = append(*testImages, testImage)
			}
		}

		return nil
	}
}
