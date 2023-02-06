package tags

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Tag store informations about single Tag
type Tag struct {
	// Name which identifies single Tag
	Name string `yaml:"name" json:"name"`
	// Value of the tag or template of it
	Value string `yaml:"value" json:"value"`
}

func NewTagFromString(val string) (Tag, error) {
	sp := strings.Split(val, "=")
	t := Tag{}

	if len(sp) > 2 {
		return t, fmt.Errorf("error parsing tag string, too many parts")
	}

	if len(sp) == 2 {
		t.Name = sp[0]
		t.Value = sp[1]
	} else {
		t.Value = val
	}

	return t, nil
}

func (t *Tag) UnmarshalYAML(value *yaml.Node) error {
	var tagTemplate string

	if err := value.Decode(&tagTemplate); err == nil {
		t.Name = "default_tag"
		t.Value = tagTemplate
		return nil
	}

	var tag map[string]string

	err := value.Decode(&tag)
	if err != nil {
		return err
	}

	t.Name = tag["name"]
	t.Value = tag["value"]

	return nil
}
