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

type ImageMap map[string]Image

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

func GetWalkFunc(resourcesDirectory string, images, testImages ImageMap, imageComponentsMap ImageToComponents) filepath.WalkFunc {
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

		AppendImagesToMap(parsedFile, images, testImages, component, imageComponentsMap)

		return nil
	}
}

func AppendImagesToMap(parsedFile ValueFile, images, testImages ImageMap, component string, components ImageToComponents) {
	for _, image := range parsedFile.Global.Images {
		// add registry info directly into the image struct
		if image.ContainerRegistryPath == "" {
			image.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
		}
		images[image.FullImageURL()] = image

		components[image.FullImageURL()] = append(components[image.FullImageURL()], component)
	}

	for _, testImage := range parsedFile.Global.TestImages {
		if testImage.ContainerRegistryPath == "" {
			testImage.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
		}
		testImages[testImage.FullImageURL()] = testImage
		components[testImage.FullImageURL()] = append(components[testImage.FullImageURL()], component)
	}
}

// GetInconsistentImages returns a list of images with the same URl but different versions or hashes
func GetInconsistentImages(images ImageMap) ImageMap {
	inconsistent := make(ImageMap)

	for imageName, image := range images {
		for image2Name, image2 := range images {
			if image.ImageURL() == image2.ImageURL() {
				inconsistent[imageName] = image
				inconsistent[image2Name] = image2
			}
		}
	}

	return inconsistent
}

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

// MergeImageMap merges images map into target one
func MergeImageMap(target ImageMap, source ImageMap) {
	for key, val := range source {
		if _, ok := target[key]; !ok {
			target[key] = val
		}
	}
}
