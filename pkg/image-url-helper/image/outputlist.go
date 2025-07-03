package image

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
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

// convertImageslist takes in a list of images and image to component mapping and creates an OutputImageList structure that can be later marshalled and used by the security scan tool
func convertImageslist(allImages ComponentImageMap) OutputImageList {
	imagesConverted := OutputImageList{}

	imageNames := make([]string, 0)
	for _, image := range allImages {
		imageNames = append(imageNames, image.Image.FullImageURL())
	}
	sort.Strings(imageNames)

	for _, fullImageURL := range imageNames {
		imageTmp := OutputImage{}
		imageTmp.Name = fullImageURL
		imageTmp.CustomFields.Image = fullImageURL

		componentNames := make([]string, 0)
		for component := range allImages[fullImageURL].Components {
			componentNames = append(componentNames, component)
		}
		imageTmp.CustomFields.Components = strings.Join(componentNames, ",")
		imagesConverted.Images = append(imagesConverted.Images, imageTmp)
	}

	return imagesConverted
}

// PrintComponentImageMap prints map of ComponentImages in a provided format
func PrintComponentImageMap(images ComponentImageMap, format string) error {
	switch strings.ToLower(format) {
	case "":
		PrintImages(images)
	case "json":
		if err := PrintImagesJSON(images); err != nil {
			return err
		}
	case "yaml":
		if err := PrintImagesYAML(images); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}

	return nil
}

// PrintImages prints otu list of images and their usage in components
func PrintImages(images ComponentImageMap) {
	imageNames := make([]string, 0)
	for _, image := range images {
		imageNames = append(imageNames, image.Image.FullImageURL())
	}
	sort.Strings(imageNames)

	for _, fullImageURL := range imageNames {
		componentNames := make([]string, 0)
		for component := range images[fullImageURL].Components {
			componentNames = append(componentNames, component)
		}
		fmt.Printf("%s, used by %s\n", fullImageURL, strings.Join(componentNames, ", "))
	}
}

// PrintImagesJSON prints JSON list with names and components for each image
func PrintImagesJSON(allImages ComponentImageMap) error {
	imagesConverted := convertImageslist(allImages)

	out, err := json.MarshalIndent(imagesConverted, "", "  ")
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(string(out))
	return nil
}

// PrintImagesYAML prints YAML list with names and components for each image
func PrintImagesYAML(allImages ComponentImageMap) error {
	imagesConverted := convertImageslist(allImages)

	out, err := yaml.Marshal(imagesConverted)
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(string(out))
	return nil
}
