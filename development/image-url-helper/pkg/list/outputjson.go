package list

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// CustomFields contains list of image custom fields
type CustomFields struct {
	Components string `json:"components" yaml:"components"`
	Image      string `json:"image" yaml:"image"`
}

// OutputImage describes an image in the output format
type OutputImage struct {
	Name         string       `json:"name" yaml:"name"`
	CustomFields CustomFields `json:"custom_fields" yaml:"custom_fields"`
}

// OutputImageList contains a list of images in the output format
type OutputImageList struct {
	Images []OutputImage `json:"images" yaml:"images"`
}

// PrintImagesJSON prints JSON list with names and components for each image
func PrintImagesJSON(images, testImages []Image, imageComponents ImageComponents) error {
	imagesConverted := convertimageslist(images, testImages, imageComponents)

	out, err := json.MarshalIndent(imagesConverted, "", "  ")
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(string(out))
	return nil
}

// PrintImagesYAML prints YAML list with names and components for each image
func PrintImagesYAML(images, testImages []Image, imageComponents ImageComponents) error {
	imagesConverted := convertimageslist(images, testImages, imageComponents)

	out, err := yaml.Marshal(imagesConverted)
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(string(out))
	return nil
}

func convertimageslist(images, testImages []Image, imageComponents ImageComponents) OutputImageList {
	imagesConverted := OutputImageList{}

	imagesCombined := images
	for _, testImage := range testImages {
		if !ImageListContains(imagesCombined, testImage) {
			imagesCombined = append(imagesCombined, testImage)
		}
	}
	sort.Slice(imagesCombined, GetSortImagesFunc(imagesCombined))

	for _, image := range imagesCombined {
		imageTmp := OutputImage{}
		imageTmp.Name = image.String()
		imageTmp.CustomFields.Image = image.String()
		components := imageComponents[image.String()]
		imageTmp.CustomFields.Components = strings.Join(components, ",")
		imagesConverted.Images = append(imagesConverted.Images, imageTmp)
	}

	return imagesConverted
}
