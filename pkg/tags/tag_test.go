package tags

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewTagFromString(t *testing.T) {
	tc := []struct {
		Name        string
		TagString   string
		ExpectedTag Tag
		ExpectErr   bool
	}{
		{
			Name:      "single tag, pass",
			TagString: "Test=Value",
			ExpectedTag: Tag{
				Name:  "Test",
				Value: "Value",
			},
		},
		{
			Name:        "malformed tag, fail",
			TagString:   "Test=Value=Other",
			ExpectedTag: Tag{},
			ExpectErr:   true,
		},
		{
			Name:        "no name, evaluate from simple value, pass",
			TagString:   "Value",
			ExpectedTag: Tag{Name: "Value", Value: "Value"},
			ExpectErr:   false,
		},
		{
			Name:        "no name, evaluate from go template, pass",
			TagString:   "v{{ .Test }}",
			ExpectedTag: Tag{Name: "vTest", Value: "v{{ .Test }}"},
			ExpectErr:   false,
		},
		{
			Name:        "no name, evaluate from go template without spaces, pass",
			TagString:   "v{{.Test}}",
			ExpectedTag: Tag{Name: "vTest", Value: "v{{.Test}}"},
			ExpectErr:   false,
		},
		{
			Name:        "no name, evaluate from go template with multiple go-templates, pass",
			TagString:   "v{{ .ShortSHA }}-{{ .Date }}",
			ExpectedTag: Tag{Name: "vShortSHA-Date", Value: "v{{ .ShortSHA }}-{{ .Date }}"},
			ExpectErr:   false,
		},
		{
			Name:        "tag template with env var",
			TagString:   `{{ .Env "PULL_BASE_REF" }}-{{ .ShortSHA }}`,
			ExpectedTag: Tag{Name: "EnvPULL_BASE_REF-ShortSHA", Value: "{{ .Env \"PULL_BASE_REF\" }}-{{ .ShortSHA }}"},
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			tg, err := NewTagFromString(c.TagString)
			if err != nil && !c.ExpectErr {
				t.Errorf("got error when no expect one: %v", err)
			}

			if !reflect.DeepEqual(tg, c.ExpectedTag) {
				t.Errorf("expected %v got %v", c.ExpectedTag, tg)
			}
		})
	}
}

func TestUnmarshallYAML(t *testing.T) {
	tc := []struct {
		Name        string
		Yaml        string
		ExpectedTag Tag
		ExpectErr   bool
	}{
		{
			Name:        "parse single value, pass",
			Yaml:        `tag-template: v{{ .Test }}`,
			ExpectedTag: Tag{Name: "default_tag", Value: "v{{ .Test }}"},
			ExpectErr:   false,
		},
		{
			Name: "parse single value, pass",
			Yaml: `tag-template:
  name: Test
  value: v{{ .Test }}`,
			ExpectedTag: Tag{Name: "Test", Value: "v{{ .Test }}"},
			ExpectErr:   false,
		},
		{
			Name:        "malformed tag, fail",
			Yaml:        `tag-template:`,
			ExpectedTag: Tag{},
			ExpectErr:   true,
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			var tagStruct struct {
				TagTemplate Tag `yaml:"tag-template" json:"tagTemplate"`
			}

			err := yaml.Unmarshal([]byte(c.Yaml), &tagStruct)

			if err != nil && !c.ExpectErr {
				t.Errorf("got unexpected error, %v", err)
			}

			if !reflect.DeepEqual(tagStruct.TagTemplate, c.ExpectedTag) {
				t.Errorf("expected %v got %v", c.ExpectedTag, tagStruct.TagTemplate)
			}
		})
	}
}
