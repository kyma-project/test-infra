package syncomponent

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSynComponent(t *testing.T) {
	// Given
	path := "/this/is/main/path/this/is/comp-name"
	mainPath := "/this/is/main/path"
	componentPath, _ := filepath.Rel(mainPath, path)

	// When
	component := NewSynComponent(componentPath, []string{"/path/to/resource"})

	// Then
	assert.Equal(t, "comp_name", component.Name)
	assert.Equal(t, "this/is/comp-name", component.Path)
	assert.Equal(t, "/path/to/resource", component.Versions[0].VersionPath)
}

func TestComponent_AddGitHashHistory(t *testing.T) {
	// Given
	key1 := subtractDaysFromToday(3)
	key2 := subtractDaysFromToday(2)
	key3 := subtractDaysFromToday(0)
	key4 := subtractDaysFromToday(1)

	history := map[int64]string{
		key1: "a1a1a1a1",
		key3: "c3c3c3c3",
		key2: "b2b2b2b2",
		key4: "d4d4d4d4",
	}

	component := NewSynComponent(".", []string{"."})

	// When
	component.GitHashHistory = history

	// Then
	assert.Equal(t, "a1a1a1a1", component.GetOldestAllowed())
}

func Test_isOutOfDate(t *testing.T) {
	t.Run("component should be out of date beacuse version is not equal to current hash and not contains in git history", func(t *testing.T) {
		component := Component{
			GitHash: "c33e5390",
			GitHashHistory: map[int64]string{
				subtractDaysFromToday(5): "4e1e6950",
			},
			Versions: []*ComponentVersions{
				{
					Version:       "b28250b7",
					ModifiedFiles: []string{"/path/to/file.go"},
				},
			},
		}
		component.CheckIsOutOfDate()
		assert.True(t, component.outOfDate)
	})

	t.Run("component should be out of date beacuse beacuse one of the version is not equal to current hash and not contains in git history", func(t *testing.T) {
		component := Component{
			GitHash: "06e635f8",
			GitHashHistory: map[int64]string{
				subtractDaysFromToday(2): "06e635f8",
				subtractDaysFromToday(6): "a00c9f2d",
			},
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
		}
		component.CheckIsOutOfDate()
		assert.True(t, component.outOfDate)
	})

	t.Run("component should not be out of date because component version contained in history logs", func(t *testing.T) {
		component := Component{
			GitHash: "86d0ddac",
			GitHashHistory: map[int64]string{
				subtractDaysFromToday(0): "86d0ddac3df8e7c4ef77f09158c68fa06566537d",
				subtractDaysFromToday(1): "b28250b72854bae58398fa599990a70d3844ec8f",
				subtractDaysFromToday(2): "c33e539019a9fca5332501098527ba05c0084973",
			},
			Versions: []*ComponentVersions{
				{
					Version:       "b28250b7",
					ModifiedFiles: []string{"/path/to/file.go"},
				},
			},
		}
		component.CheckIsOutOfDate()
		assert.False(t, component.outOfDate)
	})

	t.Run("component should not be out of date because currrent hash is equal to version", func(t *testing.T) {
		component := Component{
			GitHash: "a30ed000",
			GitHashHistory: map[int64]string{
				subtractDaysFromToday(0): "a30ed000a1b34d21d6812fc5fcd080701c042563",
				subtractDaysFromToday(1): "fc74b6ed1764c178db8ba0f78736dd9830625b37",
			},
			Versions: []*ComponentVersions{
				{
					Version:       "a30ed000",
					ModifiedFiles: []string{},
				},
			},
		}
		component.CheckIsOutOfDate()
		assert.False(t, component.outOfDate)
	})

	t.Run("component should not be out of date because only md file was changed", func(t *testing.T) {
		component := Component{
			GitHash: "8a0d7777",
			GitHashHistory: map[int64]string{
				subtractDaysFromToday(0): "6cc6a96358f654bc7a71752efec60f4fb3211c26",
				subtractDaysFromToday(1): "a00c9f2db46f422cb645b7aa8d6d0fcd2438c521",
				subtractDaysFromToday(2): "8a0d77771bcb78a253ecbd5f40ec274be79af802",
			},
			Versions: []*ComponentVersions{
				{
					Version:       "289eb8ea",
					ModifiedFiles: []string{"/path/to/file.md"},
				},
			},
		}
		component.CheckIsOutOfDate()
		assert.False(t, component.outOfDate)
	})
}

func subtractDaysFromToday(subDays int) int64 {
	if subDays > 0 {
		return time.Now().AddDate(0, 0, -(subDays)).Unix()
	}

	return time.Now().Unix()
}
