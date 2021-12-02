package common

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
