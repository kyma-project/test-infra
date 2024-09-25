package tags

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Tag store informations about single Tag
type Tag struct {
	// Name which identifies single Tag
	Name string `yaml:"name" json:"name"`
	// Value of the tag or template of it
	Value      string `yaml:"value" json:"value"`
	Validation string `yaml:"validation" json:"validation"`
}

// NewTagFromString creates new Tag from env var style string
// if not name is provided it will deduce from value based on go template syntax
func NewTagFromString(val string) (Tag, error) {
	sp := strings.Split(val, "=")
	t := Tag{}

	switch {
	case len(sp) > 2:
		return t, fmt.Errorf("error parsing tag string, too many parts")
	case len(sp) == 2:
		t.Name = sp[0]
		t.Value = sp[1]
	default:
		re := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
		t.Name = re.ReplaceAllString(val, "")

		t.Value = val
	}

	return t, nil
}

// UnmarshalYAML provides custom logic for unmarshalling tag into struct
// If not name is given it will be replaced by default_tag.
// It ensures that both use cases are supported
// TODO (dekiel): yaml config can provide a name always.
//
//	This custom unmarshaller is not needed.
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
