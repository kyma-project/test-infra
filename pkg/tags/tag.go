package tags

import (
	"fmt"
	"regexp"
	"strings"
)

// Tag store informations about single Tag
type Tag struct {
	// Name which identifies single Tag
	Name string `yaml:"name" json:"name"`
	// Value of the tag or template of it
	Value string `yaml:"value" json:"value"`
	// Validation is a regex pattern to validate the tag value after it has been parsed
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
