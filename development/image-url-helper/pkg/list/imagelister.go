package list

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

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

// String returns complete image URL with version or SHA
func (i Image) String() string {
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

func GetWalkFunc(resourcesDirectory string, images, testImages *[]Image, imageComponents ImageComponents) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
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

		var parsedFile ValueFile

		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(yamlFile, &parsedFile)
		if err != nil {
			return err
		}

		component := strings.Replace(path, resourcesDirectory+"/", "", -1)
		component = strings.Replace(component, "/values.yaml", "", -1)

		for _, image := range parsedFile.Global.Images {
			// add registry info directly into the image struct
			if image.ContainerRegistryPath == "" {
				image.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
			}
			// remove duplicates
			if !ImageListContains(*images, image) {
				*images = append(*images, image)
			}
			imageComponents[image.String()] = append(imageComponents[image.String()], component)
		}

		for _, testImage := range parsedFile.Global.TestImages {
			if testImage.ContainerRegistryPath == "" {
				testImage.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
			}
			if !ImageListContains(*testImages, testImage) {
				*testImages = append(*testImages, testImage)
			}
			imageComponents[testImage.String()] = append(imageComponents[testImage.String()], component)
		}

		return nil
	}
}

// RemoveDoubles removes all duplicates
func RemoveDoubles(images []Image) []Image {
	var dedupedImages []Image
	for _, image := range images {
		exists := false
		for _, deduped := range dedupedImages {
			if image == deduped {
				exists = true
			}
		}
		if !exists {
			dedupedImages = append(dedupedImages, image)
		}
	}
	return dedupedImages
}

// GetInconsistentImages returns a list of images with the same URl but different versions or hashes
func GetInconsistentImages(images []Image) []Image {
	var inconsistent []Image
	hasDoubles := make(map[string][]Image)

	for _, image := range images {
		hasDoubles[image.ImageURL()] = append(hasDoubles[image.ImageURL()], image)
	}

	for _, images := range hasDoubles {
		if len(images) > 1 {
			inconsistent = append(inconsistent, images...)
		}
	}

	sort.Slice(inconsistent, GetSortImagesFunc(inconsistent))
	return inconsistent
}

// PrintImages prints otu list of images and their usage in components
func PrintImages(images []Image, imageComponents ImageComponents) {
	sort.Slice(images, GetSortImagesFunc(images))
	for _, image := range images {
		components := imageComponents[image.String()]
		fmt.Printf("%s, used by %s\n", image, strings.Join(components, ", "))
	}
}
