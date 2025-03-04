package extractimageurls

import (
	"io"

	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

// FromKubernetesDeployments returns list of images found in provided file
func FromKubernetesDeployments(reader io.Reader) ([]string, error) {
	var images []string

	// Read all data from reader
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Split file into sections
	sections := SplitYamlIntoSections(data)

	for _, section := range sections {
		var file v1.Deployment
		err := yaml.Unmarshal(section, &file)
		if err != nil {
			return nil, err
		}

		images = append(images, extractImageUrlsFromStruct(file)...)
	}
	return images, nil
}

// extractImageUrlsFromStruct extracts image urls from given struct and returns it as slice of strings
func extractImageUrlsFromStruct(file v1.Deployment) []string {
	images := []string{}
	for _, image := range file.Spec.Template.Spec.Containers {
		images = append(images, image.Image)
	}

	return images
}
# (2025-03-04)