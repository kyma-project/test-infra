package tools

import (
	"fmt"
	"io/ioutil"

	sc "github.com/kyma-project/test-infra/development/tools/pkg/synchronizer/syncomponent"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const versionFile = "values.yaml"

// Values represents file for chart resource values
type Values struct {
	Global Components `yaml:"global"`
}

// Components represents list of components in global key
type Components struct {
	element map[string]Component
}

// Component represents single component in global key
type Component struct {
	Version string `yaml:"version"`
}

// UnmarshalYAML helper function to unamrchal compnents with different key names
func (v *Components) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var components map[string]Component
	if err := unmarshal(&components); err != nil {
		if _, ok := err.(*yaml.TypeError); !ok {
			return err
		}
	}
	v.element = components

	return nil
}

// FindComponentVersion sets versions for component based on values.yaml file
func FindComponentVersion(rootDir string, component *sc.Component) error {
	for _, version := range component.Versions {
		yamlContent, err := valueYamlContent(rootDir, version.VersionPath, versionFile)
		versionValue, err := findVersion(yamlContent, component.Name)
		if err != nil {
			return errors.Wrapf(err, "while find component version in %s/%s", version.VersionPath, versionFile)
		}
		version.Version = versionValue
	}

	return nil
}

func findVersion(yamlContent []byte, name string) (string, error) {
	var val Values
	err := yaml.Unmarshal(yamlContent, &val)
	if err != nil {
		return "", errors.Wrapf(err, "while unmarshal file %s", versionFile)
	}

	for componentName, componentValue := range val.Global.element {
		if name == componentName {
			return componentValue.Version, nil
		}
	}

	return "", errors.Errorf("component %q not exist for the indicated version file path", name)
}

func valueYamlContent(rootDir, path, file string) ([]byte, error) {
	pathFile := fmt.Sprintf("%s/%s/%s", rootDir, path, file)
	content, err := ioutil.ReadFile(pathFile)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "while reading file %q", pathFile)
	}

	return content, nil
}
