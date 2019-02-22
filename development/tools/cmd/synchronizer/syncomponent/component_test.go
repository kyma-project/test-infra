package syncomponent

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_RelativePathToComponent(t *testing.T) {
	// Given
	path := "/this/is/main/path/this/is/subpath"

	// When
	result := RelativePathToComponent("/this/is/main/path", path)

	// Then
	if result != "this/is/subpath" {
		t.Fatalf("Subpath to component is incorrect: %q", result)
	}
}

func Test_daysDelta(t *testing.T) {
	ts := []struct {
		unixString string
		unixTime   int64
		expected   int
	}{
		{unixTimeAsString(3), time.Now().Unix(), 3},
		{unixTimeAsString(2), time.Now().Unix(), 2},
		{unixTimeAsString(0), time.Now().Unix(), 0},
		{unixTimeAsString(6), time.Now().Unix(), 6},
	}

	for _, tu := range ts {
		assert.Equal(t, daysDelta(tu.unixString, tu.unixTime), tu.expected)
	}
}

func Test_versionIsOutOfDate(t *testing.T) {
	// Given
	tsFalse := map[string]ComponentVersions{
		"bc120200aa3d0d42b1vca21d4f13fa9853529ec0": {Version: "bc120200"},
		"cef7f8fe1ac294342f8450432945674a86ea7da3": {Version: "cef7f8fe1ac2"},
		"e918d861abc3e51618f6371efde1cb6d2748b139": {Version: "e918d8"},
	}

	tsFilesFalse := map[string]ComponentVersions{
		"bc120200aa3d0d42b1vca21d4f13fa9853529ec0": {Version: "e918d861", ModifiedFiles: []string{"/path/to/file.md"}},
		"cef7f8fe1ac294342f8450432945674a86ea7da3": {Version: "e918d861", ModifiedFiles: []string{"/path/to/file.svg"}},
		"e918d861abc3e51618f6371efde1cb6d2748b139": {Version: "cef7f8fe", ModifiedFiles: []string{"/path/to/file.png", "/path/to/file.dia"}},
	}

	tsTrue := map[string]ComponentVersions{
		"bc120200aa3d0d42b1vca21d4f13fa9853529ec0": {Version: "e918d8", ModifiedFiles: []string{"/path/to/file.go"}},
		"cef7f8fe1ac294342f8450432945674a86ea7da3": {Version: "bc120200", ModifiedFiles: []string{"/path/to/file.go"}},
		"b28250b71cc450e34d7d5fb98f5cda2de0195e3e": {Version: "b000f7b7", ModifiedFiles: []string{"/path/to/file.go"}},
		"e918d861abc3e51618f6371efde1cb6d2748b139": {Version: "cef7f8fe1ac2", ModifiedFiles: []string{"/path/to/file.go"}},
		"bc7646885aa1bb54d6ebebc31db9c196456cb42c": {Version: "b28250b7", ModifiedFiles: []string{"/path/to/file"}},
		"b000f7b79b72f8c70cbf6b2a9552049be32c7b21": {Version: "bc764688", ModifiedFiles: []string{"/path/to/file.md", "/path/to/file.go"}},
	}

	// Then
	for hash, cv := range tsFalse {
		cv.setOutOfDate(hash)
		assert.False(t, cv.outOfDate)
	}
	for hash, cv := range tsFilesFalse {
		cv.setOutOfDate(hash)
		assert.False(t, cv.outOfDate)
	}
	for hash, cv := range tsTrue {
		cv.setOutOfDate(hash)
		assert.True(t, cv.outOfDate)
	}
}

func Test_isOutOfDate(t *testing.T) {
	// Given
	tsTrue := []Component{
		{
			GitHash:    "bc120200aa3d0d42b1vca21d4f13fa9853529ec0",
			CommitDate: unixTimeAsString((defaultOutOfDateDays + 1)),
			Versions: []*ComponentVersions{
				{
					Version:       "b28250b7",
					ModifiedFiles: []string{"/path/to/file.go"},
				},
			},
		},
		{
			GitHash:    "06e635f865fc473049e453245e2122dc8f17c532",
			CommitDate: unixTimeAsString((defaultOutOfDateDays + 1)),
			Versions: []*ComponentVersions{
				{
					Version:       "06e635f8",
					ModifiedFiles: []string{},
				},
				{
					Version:       "b28250b7",
					ModifiedFiles: []string{"/path/to/file"},
				},
			},
		},
	}

	tsFalse := []Component{
		{
			// is not out of date because component not achive expired date
			GitHash:    "86d0ddac3df8e7c4ef77f09158c68fa06566537d",
			CommitDate: unixTimeAsString(defaultOutOfDateDays),
			Versions: []*ComponentVersions{
				{
					Version:       "b28250b7",
					ModifiedFiles: []string{"/path/to/file.go"},
				},
			},
		},
		{
			// is not out of date because hashes are equal
			GitHash:    "a30ed000cb1fe91640376f6d6aac64bf1e2bd080",
			CommitDate: unixTimeAsString((defaultOutOfDateDays + 3)),
			Versions: []*ComponentVersions{
				{
					Version:       "a30ed000",
					ModifiedFiles: []string{},
				},
			},
		},
		{
			//is not out of date because only md file was changed
			GitHash:    "289eb8ea3ec73367581ae3a6860981b771495b83",
			CommitDate: unixTimeAsString((defaultOutOfDateDays + 3)),
			Versions: []*ComponentVersions{
				{
					Version:       "289eb8ea",
					ModifiedFiles: []string{"/path/to/file.md"},
				},
			},
		},
	}

	// Then
	for _, comp := range tsTrue {
		comp.SetOutOfDate(defaultOutOfDateDays)
		assert.True(t, comp.outOfDate)
	}
	for _, comp := range tsFalse {
		comp.SetOutOfDate(defaultOutOfDateDays)
		assert.False(t, comp.outOfDate, fmt.Sprintf("Component with hash %q", comp.GitHash))
	}
}

func unixTimeAsString(subDays int) string {
	if subDays > 0 {
		return strconv.FormatInt(time.Now().AddDate(0, 0, -(subDays)).Unix(), 10)
	}

	return strconv.FormatInt(time.Now().Unix(), 10)
}
