package tools

import (
	"testing"

	sc "github.com/kyma-project/test-infra/development/tools/pkg/synchronizer/syncomponent"
	"github.com/stretchr/testify/assert"
)

var testYaml = `
irrelevant_one: 
  some_key: 
    some_sub_key: some_value

irrelevant_two: 
  some_key: 
    some_sub_key: some_value

global:
  other_bool_value: false
  some_bool_value: false
  wrong_struct_of_component: 
    wrong_key: some_value
  component_one: 
    wrong_key: val_one
    wrong_value: val_two
  component_two: 
    dir: develop/
    version: 1a1a1a1a
  component_three: 
    dir: develop/
    version: 2b2b2b2b

irrelevant_three: 
  some_key: 
    some_sub_key: 
      some_other_key: 
        - 
          value_one: one
          value_three: three
          value_two: two
`

var wrongYaml = `
irrelevant_one: 
  some_key: 
    some_sub_key: some_value

irrelevant_two: 
  some_key: 
    some_sub_key: some_value
`

func TestFindVersion(t *testing.T) {
	// Given
	ts := sc.Component{
		Name: "component_two",
		Versions: []*sc.ComponentVersions{
			{VersionPath: "."},
			{VersionPath: "."},
		},
	}

	// When
	ver, err := findVersion([]byte(testYaml), ts.Name)

	// Then
	assert.Equal(t, "1a1a1a1a", ver)
	assert.Empty(t, err)

	// When
	verEmpty, err := findVersion([]byte(wrongYaml), ts.Name)

	// Then
	assert.Equal(t, "", verEmpty)
	assert.NotEmpty(t, err)
}
