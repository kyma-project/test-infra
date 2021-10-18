package main

import (
	"encoding/json"
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

type ImageComponents map[string][]string

func main() {
	// map image name to list of components that are using it
	imageComponents := make(ImageComponents)

	flag.Parse()
	fmt.Printf("Looking for images in \"%s\"\n\n", *kymaResourcesDirectory)

	var images []imagelister.Image
	var testImages []imagelister.Image

	err := filepath.Walk(*kymaResourcesDirectory, getWalkFunc(&images, &testImages, imageComponents))
	if err != nil {
		fmt.Printf("Cannot traverse directory: %s\n", err)
		os.Exit(2)
	}

	sort.Slice(images, imagelister.GetSortImagesFunc(images))

	fmt.Println("images:")
	// TODO function here
	for _, image := range images {
		components := imageComponents[image.String()]
		fmt.Printf("%s, used by %s\n", image, strings.Join(components, ", "))
	}

	sort.Slice(testImages, imagelister.GetSortImagesFunc(testImages))

	fmt.Println("test images:")
	// TODO function here
	for _, testImage := range testImages {
		components := imageComponents[testImage.String()]
		fmt.Printf("%s, used by %s\n", testImage, strings.Join(components, ", "))
	}

	if *outputJSON != "" {
		err = writeImagesJSON(images, testImages, imageComponents)
		if err != nil {
			fmt.Printf("Cannot save JSON: %s\n", err)
			os.Exit(2)
		}
	}
}

func writeImagesJSON(images, testImages []imagelister.Image, imageComponents ImageComponents) error {
	imagesCombined := images
	for _, testImage := range testImages {
		if !imagelister.ImageListContains(imagesCombined, testImage) {
			imagesCombined = append(imagesCombined, testImage)
		}
	}
	sort.Slice(imagesCombined, imagelister.GetSortImagesFunc(imagesCombined))

	// TODO convert images
	imagesConverted := imagelister.ImagesJSON{}
	for _, image := range imagesCombined {
		imageTmp := imagelister.ImageJSON{}
		imageTmp.Name = image.String()
		imageTmp.CustomFields.Image = image.String()
		components := imageComponents[image.String()]
		imageTmp.CustomFields.Components = strings.Join(components, ",")
		imagesConverted.Images = append(imagesConverted.Images, imageTmp)
	}

	outputFile, err := os.Create(*outputJSON)
	if err != nil {
		return fmt.Errorf("error creating output file: %s", err)
	}
	defer outputFile.Close()

	out, err := json.MarshalIndent(imagesConverted, "", "  ")
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	outputFile.Write(out)
	return nil
}

func getWalkFunc(images, testImages *[]imagelister.Image, imageComponents ImageComponents) filepath.WalkFunc {
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
