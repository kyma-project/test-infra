package image

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

func AppendImagesToMap(parsedFile ValueFile, images, testImages ComponentImageMap, component string) {
	for _, image := range parsedFile.Global.Images {
		// add registry info directly into the image struct
		if image.ContainerRegistryURL == "" {
			image.ContainerRegistryURL = parsedFile.Global.ContainerRegistry.Path
		}

		if _, ok := images[image.FullImageURL()]; ok {
			images[image.FullImageURL()].Components[component] = true
		} else {
			images[image.FullImageURL()] = ComponentImage{
				Components: map[string]bool{component: true},
				Image:      image,
			}
		}
	}

	for _, testImage := range parsedFile.Global.TestImages {
		if testImage.ContainerRegistryURL == "" {
			testImage.ContainerRegistryURL = parsedFile.Global.ContainerRegistry.Path
		}

		if _, ok := testImages[testImage.FullImageURL()]; ok {
			testImages[testImage.FullImageURL()].Components[component] = true
		} else {
			testImages[testImage.FullImageURL()] = ComponentImage{
				Components: map[string]bool{component: true},
				Image:      testImage,
			}
		}
	}
}
