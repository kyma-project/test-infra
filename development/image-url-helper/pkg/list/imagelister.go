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

type ImageList []Image

// ImageToComponents is a map that for each image name stores list of components that are using this image
type ImageToComponents map[string][]string

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

// ImageListContains checks if list of images contains already the same image
func ImageListContains(list ImageList, image Image) bool {
	for _, singleImage := range list {
		if singleImage == image {
			return true
		}
	}
	return false
}

// Len return length of a list
func (images ImageList) Len() int {
	return len(images)
}

// Less returns if the one image in the list should be before the ther one
func (images ImageList) Less(i, j int) bool {
	return images[i].FullImageURL() < images[j].FullImageURL()
}

// Swap swaps two images in the list
func (images ImageList) Swap(i, j int) {
	images[i], images[j] = images[j], images[i]
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

func GetWalkFunc(resourcesDirectory string, images, testImages *ImageList, imageComponentsMap ImageToComponents) filepath.WalkFunc {
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

		AppendImagesToList(parsedFile, images, testImages, component, imageComponentsMap)

		return nil
	}
}

func AppendImagesToList(parsedFile ValueFile, images, testImages *ImageList, component string, components ImageToComponents) {
	for _, image := range parsedFile.Global.Images {
		// add registry info directly into the image struct
		if image.ContainerRegistryPath == "" {
			image.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
		}
		// remove duplicates
		if !ImageListContains(*images, image) {
			*images = append(*images, image)
		}
		components[image.FullImageURL()] = append(components[image.FullImageURL()], component)
	}

	for _, testImage := range parsedFile.Global.TestImages {
		if testImage.ContainerRegistryPath == "" {
			testImage.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
		}
		if !ImageListContains(*testImages, testImage) {
			*testImages = append(*testImages, testImage)
		}
		components[testImage.FullImageURL()] = append(components[testImage.FullImageURL()], component)
	}
}

// RemoveDoubles removes all duplicates
func RemoveDoubles(images ImageList) ImageList {
	var dedupedImages ImageList
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
func GetInconsistentImages(images ImageList) ImageList {
	var inconsistent ImageList
	hasDoubles := make(map[string]ImageList)

	for _, image := range images {
		hasDoubles[image.ImageURL()] = append(hasDoubles[image.ImageURL()], image)
	}

	for _, images := range hasDoubles {
		if len(images) > 1 {
			inconsistent = append(inconsistent, images...)
		}
	}

	return inconsistent
}

// PrintImages prints otu list of images and their usage in components
func PrintImages(images ImageList, imageComponentsMap ImageToComponents) {
	sort.Sort(images)
	for _, image := range images {
		components := imageComponentsMap[image.FullImageURL()]
		fmt.Printf("%s, used by %s\n", image.FullImageURL(), strings.Join(components, ", "))
	}
}
