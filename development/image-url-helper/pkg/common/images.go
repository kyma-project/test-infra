package common

import (
	"fmt"
)

type ImageMap map[string]Image

// Image contains info about a singular image
type Image struct {
	ContainerRegistryURL    string `yaml:"containerRegistryPath,omitempty"`
	ContainerRepositoryPath string `yaml:"directory,omitempty"`
	Name                    string `yaml:"name,omitempty"`
	Version                 string `yaml:"version,omitempty"`
	SHA                     string `yaml:"sha,omitempty"`
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
func GetInconsistentImages(images ImageMap) ImageMap {
	inconsistent := make(ImageMap)
	tmpImages := make(map[string]string)

	for imageName, image := range images {
		if tmpImageName, ok := tmpImages[image.ImageURL()]; ok {
			if imageName != tmpImageName {
				inconsistent[imageName] = image
				inconsistent[tmpImageName] = images[tmpImageName]
			}
		} else {
			tmpImages[image.ImageURL()] = imageName
		}
	}

	return inconsistent
}

// MergeImageMap merges images map into target one
func MergeImageMap(target ImageMap, source ImageMap) {
	for key, val := range source {
		if _, ok := target[key]; !ok {
			target[key] = val
		}
	}
}
