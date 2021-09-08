package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type BuildConfig struct {
	Concurrent bool `yaml:"concurrent" json:"concurrent"`
	Steps      []struct {
		Name   string  `yaml:"name" json:"name"`
		Images []Image `yaml:"images" json:"images"`
	} `yaml:"steps" json:"steps"`
}

type Image struct {
	Name         string `yaml:"name" json:"name"`
	Tag          string `yaml:"tag" json:"tag"`
	Context      string `yaml:"context" json:"context"`
	Dockerfile   string `yaml:"dockerfile" json:"dockerfile"`
	RemotePrefix string `yaml:"remotePrefix" json:"remotePrefix"`
}

func (i *Image) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tmp Image
	var p tmp
	if err := unmarshal(&p); err != nil {
		return err
	}

	*i = Image(p)
	return parseImage(i)
}

func parseImage(i *Image) error {
	if len(i.Name) == 0 {
		return fmt.Errorf("image name cannot be empty")
	}
	if len(i.Context) == 0 {
		return fmt.Errorf("image %s: context cannot be empty", i.Name)
	}
	if len(i.Dockerfile) == 0 {
		i.Dockerfile = "Dockerfile"
	}
	return nil
}

func NewFromFile(file string) (*BuildConfig, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// TODO (@Ressetkk): Separate function for parsing file type from reader
	var c BuildConfig
	err = yaml.NewDecoder(f).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
