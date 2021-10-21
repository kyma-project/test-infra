package list

import "fmt"

// SortFunction is a function type used in slice sorting
type SortFunction func(i, j int) bool

// ImageComponents is a map that for each image name stores list of components that are using this image
type ImageComponents map[string][]string

// Image contains info about a singular image
type Image struct {
	ContainerRegistryPath string `yaml:"containerRegistryPath,omitempty"`
	Directory             string `yaml:"directory,omitempty"`
	Name                  string `yaml:"name,omitempty"`
	Version               string `yaml:"version,omitempty"`
	SHA                   string `yaml:"sha,omitempty"`
}

// String returns complete image URL
func (i Image) String() string {
	registry := i.ContainerRegistryPath
	if i.Directory != "" {
		registry += "/" + i.Directory
	}
	version := ":" + i.Version
	if i.SHA != "" {
		version = "@sha256:" + i.SHA
	}
	return fmt.Sprintf("%s/%s%s", registry, i.Name, version)
}

// ImageListContains checks if list of images contains already the same image
func ImageListContains(list []Image, image Image) bool {
	for _, singleImage := range list {
		if singleImage == image {
			return true
		}
	}
	return false
}

// GetSortImagesFunc returns sorting function for images list
func GetSortImagesFunc(images []Image) SortFunction {
	return func(i, j int) bool {
		return images[i].String() < images[j].String()
	}
}

// ContainerRegistry stores path to a container registry
type ContainerRegistry struct {
	Path string `yaml:"path,omitempty"`
}

// GlobalKey contains all keys with image info inside global key
type GlobalKey struct {
	ContainerRegistry ContainerRegistry `yaml:"containerRegistry,omitempty"`
	Images            map[string]Image  `yaml:"images,omitempty"`
	TestImages        map[string]Image  `yaml:"testImages,omitempty"`
}

// ValueFile provides simple mapping to a values.yaml file
type ValueFile struct {
	Global GlobalKey `yaml:"global,omitempty"`
}
