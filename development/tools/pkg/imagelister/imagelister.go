package imagelister

import "fmt"

type sortFunction func(i, j int) bool

type Image struct {
	ContainerRegistryPath string `yaml:"containerRegistryPath,omitempty"`
	Directory             string `yaml:"directory,omitempty"`
	Name                  string `yaml:"name,omitempty"`
	Version               string `yaml:"version,omitempty"`
	SHA                   string `yaml:"sha,omitempty"`
}

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

func ImageListContains(list []Image, image Image) bool {
	for _, singleImage := range list {
		if singleImage == image {
			return true
		}
	}
	return false
}

type containerRegistry struct {
	Path string `yaml:"path,omitempty"`
}

func GetSortImagesFunc(images []Image) sortFunction {
	return func(i, j int) bool {
		return images[i].String() < images[j].String()
	}
}

type globalKey struct {
	ContainerRegistry containerRegistry `yaml:"containerRegistry,omitempty"`
	Images            map[string]Image  `yaml:"images,omitempty"`
	TestImages        map[string]Image  `yaml:"testImages,omitempty"`
}

type ValueFile struct {
	Global globalKey `yaml:"global,omitempty"`
}
