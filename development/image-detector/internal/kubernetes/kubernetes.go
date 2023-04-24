package kubernetes

import (
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type DeploymentFile struct {
	Spec SpecField `yaml:"spec"`
}

type SpecField struct {
	Template Template `yaml:"template"`
}

type Template struct {
	Spec PodSpec `yaml:"spec"`
}

type PodSpec struct {
	Containers []Container `yaml:"containers"`
}

type Container struct {
	Image string `yaml:"image"`
}

func Extract(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return extract(f)
}

func extract(reader io.Reader) ([]string, error) {
	images := []string{}

	decoder := yaml.NewDecoder(reader)
	for {
		var file DeploymentFile
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

func extractImagesFromStruct(file DeploymentFile) []string {
	images := []string{}
	for _, image := range file.Spec.Template.Spec.Containers {
		images = append(images, image.Image)
	}

	return images
}
