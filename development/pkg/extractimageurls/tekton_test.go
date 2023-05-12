package extractimageurls

import (
	"reflect"
	"strings"
	"testing"
)

func TestFromTektonTask(t *testing.T) {
	tc := []struct {
		Name           string
		ExpectedImages []string
		WantErr        bool
		FileContent    string
	}{}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			images, err := FromTektonTask(strings.NewReader(c.FileContent))
			if err != nil && !c.WantErr {
				t.Errorf("error occurred but not expected: %s", err)
			}

			if !reflect.DeepEqual(images, c.ExpectedImages) {
				t.Errorf("FromTektonTask(): Got %v, but expected %v", images, c.ExpectedImages)
			}
		})
	}
}
