package common

import (
	"fmt"
	"sort"
	"strings"
)

// ImageToComponents is a map that for each image name stores list of components that are using this image
type ImageToComponents map[string][]string

// PrintImages prints otu list of images and their usage in components
func PrintImages(images ImageMap, imageComponentsMap ImageToComponents) {
	imageNames := make([]string, 0)
	for _, image := range images {
		imageNames = append(imageNames, image.FullImageURL())
	}
	sort.Strings(imageNames)

	for _, fullImageURL := range imageNames {
		components := imageComponentsMap[fullImageURL]
		fmt.Printf("%s, used by %s\n", fullImageURL, strings.Join(components, ", "))
	}
}
