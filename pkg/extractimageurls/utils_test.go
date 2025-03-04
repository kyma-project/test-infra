package extractimageurls

import (
	"reflect"
	"testing"
)

func TestUniqueImages(t *testing.T) {
	tc := []struct {
		Name           string
		GivenImages    []string
		ExpectedImages []string
	}{
		{
			Name:           "remove duplicated images",
			GivenImages:    []string{"same/image:test", "same/image:test"},
			ExpectedImages: []string{"same/image:test"},
		},
		{
			Name:           "keep images for unique list",
			GivenImages:    []string{"unique/image:123", "other-unique/image:123"},
			ExpectedImages: []string{"unique/image:123", "other-unique/image:123"},
		},
		{
			Name:           "find multiple duplicates in long images list",
			GivenImages:    []string{"some/image:test", "other/image:test", "other/image:test", "some-other/image:test", "one-more/image:test-tag", "yet/another-image:other-tag", "yet/another-image:other-tag", "yet/another-image:tag", "yet/another-image:other-tag"},
			ExpectedImages: []string{"some/image:test", "other/image:test", "some-other/image:test", "one-more/image:test-tag", "yet/another-image:other-tag", "yet/another-image:tag"},
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			actual := UniqueImages(c.GivenImages)

			if !reflect.DeepEqual(actual, c.ExpectedImages) {
				t.Errorf("UniqueImages(): Got %v, but expected %v", actual, c.ExpectedImages)
			}
		})
	}
}
# (2025-03-04)