package extractimageurls

import (
	"io"
	"regexp"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"sigs.k8s.io/yaml"
)

// FromTektonTask extracts list of image urls from tekton task definition
func FromTektonTask(reader io.Reader) ([]string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var task v1beta1.Task
	err = yaml.UnmarshalStrict(data, &task)
	if err != nil {
		return nil, err
	}

	var images []string

	for _, step := range task.Spec.Steps {
		images = extractImageURLFromTaskStep(step, task.Spec.Params)
	}

	return images, nil
}

// extractImageURLFromTaskStep returns list of images used in tekton task steps
// If image is provided via parameter, it will return default value for that parameter as image
func extractImageURLFromTaskStep(step v1beta1.Step, params []v1beta1.ParamSpec) []string {
	re := regexp.MustCompile(`\$\(params.([^)]+)\)`)

	var images []string

	paramNames := re.FindAllStringSubmatch(step.Image, -1)
	if len(paramNames) < 1 {
		return []string{step.Image}
	}

	for _, paramName := range paramNames {
		for _, param := range params {
			// Ignore if not param name
			if len(paramName) < 1 {
				continue
			}

			// Ignore if not default value
			if param.Default == nil {
				continue
			}

			if param.Name == paramName[1] {
				images = append(images, param.Default.StringVal)
			}
		}
	}

	return images
}
