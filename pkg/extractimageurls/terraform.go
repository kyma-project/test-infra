package extractimageurls

import (
	"io"
	"regexp"
)

// FromTerraform extracts docker images from terraform files.
// It receives reader with terraform file content and returns list of images or error if any.
func FromTerraform(reader io.Reader) ([]string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`([a-z0-9]+(?:[.-][a-z0-9]+)*/)*([a-z0-9]+(?:[.-][a-z0-9]+)*)(?::[a-z0-9.-]+)?/([a-z0-9-]+)/([a-z0-9-]+)(?::[a-z0-9.-]+)`)
	substrings := re.FindAllStringSubmatch(string(data), -1)

	var images []string
	for _, substr := range substrings {
		images = append(images, substr[0])
	}

	return images, nil
}
# (2025-03-04)