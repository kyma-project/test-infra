package kubernetes

import (
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
	var file DeploymentFile
	err := yaml.NewDecoder(reader).Decode(&file)
	if err != nil {
		return nil, err
	}

	images := []string{}
	for _, image := range file.Spec.Template.Spec.Containers {
		images = append(images, image.Image)
	}

	return images, nil
}
