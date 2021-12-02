package common

import (
	"fmt"
)

type ImageMap map[string]Image

// Image contains info about a singular image
type Image struct {
	ContainerRegistryPath string `yaml:"containerRegistryPath,omitempty"`
	Directory             string `yaml:"directory,omitempty"`
	Name                  string `yaml:"name,omitempty"`
	Version               string `yaml:"version,omitempty"`
	SHA                   string `yaml:"sha,omitempty"`
}

// FullImageURL returns complete image URL with version or SHA
func (i Image) FullImageURL() string {
	version := ":" + i.Version
	if i.SHA != "" {
		version = "@sha256:" + i.SHA
	}
	return fmt.Sprintf("%s%s", i.ImageURL(), version)
}

// ImageURL returns image URL without version
func (i Image) ImageURL() string {
	registry := i.ContainerRegistryPath
	if i.Directory != "" {
		registry += "/" + i.Directory
	}
	return fmt.Sprintf("%s/%s", registry, i.Name)
}

// GetInconsistentImages returns a list of images with the same URl but different versions or hashes
func GetInconsistentImages(images ImageMap) ImageMap {
	inconsistent := make(ImageMap)

	for imageName, image := range images {
		for image2Name, image2 := range images {
			if image.ImageURL() == image2.ImageURL() && imageName != image2Name {
				inconsistent[imageName] = image
				inconsistent[image2Name] = image2
			}
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
