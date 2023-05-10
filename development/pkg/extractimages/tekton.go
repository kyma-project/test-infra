package extractimages

import (
	"io"
)

type Step struct {
	Image string `yaml:"image"`
}

type Spec struct {
	Steps []Step `yaml:"steps"`
}

type Task struct {
	Spec Spec `yaml:spec"`
}

func FromTektonTask(reader io.Reader) ([]string, error) {}
