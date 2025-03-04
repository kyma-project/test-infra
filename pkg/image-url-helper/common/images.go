package common

import (
	"fmt"
)

// ComponentImageMap defines type for map of ComponentImages, where key is a full image URL
type ComponentImageMap map[string]ComponentImage

// Image contains info about a singular image
type Image struct {
	ContainerRegistryURL    string `yaml:"containerRegistryPath,omitempty"`
	ContainerRepositoryPath string `yaml:"directory,omitempty"`
	Name                    string `yaml:"name,omitempty"`
	Version                 string `yaml:"version,omitempty"`
	SHA                     string `yaml:"sha,omitempty"`
}

// ComponentImage contains image and map of components using this image
type ComponentImage struct {
	Components map[string]bool
	Image      Image
}

// FullImageURL returns complete image URL with version or SHA
func (i Image) FullImageURL() string {
	version := ""
	if i.SHA != "" {
		version = "@sha256:" + i.SHA
	} else {
		version = ":" + i.Version
	}
	return fmt.Sprintf("%s%s", i.ImageURL(), version)
}

// ImageURL returns image URL without version
func (i Image) ImageURL() string {
	registry := i.ContainerRegistryURL
	if i.ContainerRepositoryPath != "" {
		registry += "/" + i.ContainerRepositoryPath
	}
	return fmt.Sprintf("%s/%s", registry, i.Name)
}

// GetInconsistentImages returns a list of images with the same URl but different versions or hashes
func GetInconsistentImages(images ComponentImageMap) ComponentImageMap {
	inconsistent := make(ComponentImageMap)
	tmpImages := make(map[string]string)

	for imageName, image := range images {
		if tmpImageName, ok := tmpImages[image.Image.ImageURL()]; ok {
			if imageName != tmpImageName {
				inconsistent[imageName] = image
				inconsistent[tmpImageName] = images[tmpImageName]
			}
		} else {
			tmpImages[image.Image.ImageURL()] = imageName
		}
	}

	return inconsistent
}

// MergeImageMap merges images map into target one
func MergeImageMap(target ComponentImageMap, source ComponentImageMap) {
	for key := range source {
		if _, ok := target[key]; !ok {
			// we have the same image in both maps
			// TODO
			tmpComponentImage := ComponentImage{Image: source[key].Image, Components: source[key].Components}

			// join both maps
			for component := range target[key].Components {
				tmpComponentImage.Components[component] = true
			}

			target[key] = tmpComponentImage
		}
	}
}
# (2025-03-04)