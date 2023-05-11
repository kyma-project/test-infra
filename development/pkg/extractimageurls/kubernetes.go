package extractimageurls

import (
	"errors"
	"io"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/apps/v1"
)

// FromKubernetesDeployments returns list of images found in provided file
func FromKubernetesDeployments(reader io.Reader) ([]string, error) {
	images := []string{}

	decoder := yaml.NewDecoder(reader)
	for {
		var file v1.Deployment
		err := decoder.Decode(&file)

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		images = append(images, extractImagesFromStruct(file)...)
	}
	return images, nil
}

func extractImagesFromStruct(file v1.Deployment) []string {
	images := []string{}
	for _, image := range file.Spec.Template.Spec.Containers {
		images = append(images, image.Image)
	}

	return images
}
