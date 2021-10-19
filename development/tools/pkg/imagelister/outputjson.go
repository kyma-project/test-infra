package imagelister

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
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

// WriteImagesJSON exports JSON list with names and components for each image
func WriteImagesJSON(outputFilename string, images, testImages []Image, imageComponents ImageComponents) error {
	imagesCombined := images
	for _, testImage := range testImages {
		if !ImageListContains(imagesCombined, testImage) {
			imagesCombined = append(imagesCombined, testImage)
		}
	}
	sort.Slice(imagesCombined, GetSortImagesFunc(imagesCombined))

	// TODO convert images
	imagesConverted := OutputImageList{}
	for _, image := range imagesCombined {
		imageTmp := OutputImage{}
		imageTmp.Name = image.String()
		imageTmp.CustomFields.Image = image.String()
		components := imageComponents[image.String()]
		imageTmp.CustomFields.Components = strings.Join(components, ",")
		imagesConverted.Images = append(imagesConverted.Images, imageTmp)
	}

	outputFile, err := os.Create(outputFilename)
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
