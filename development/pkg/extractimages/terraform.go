package extractimages

import (
	"io"
	"os"
	"regexp"
)

// FromTerraform extracts docker images from terraform files.
// It receives path to single terraform file and returns list of images or error if any.
func FromTerraform(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return extractDockerImagesFromTerraform(f)
}

func extractDockerImagesFromTerraform(reader io.Reader) ([]string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`([a-z0-9]+(?:[.-][a-z0-9]+)*/)*([a-z0-9]+(?:[.-][a-z0-9]+)*)(?::[a-z0-9.-]+)?/([a-z0-9-]+)/([a-z0-9-]+)(?::[a-z0-9.-]+)?`)
	substrings := re.FindAllStringSubmatch(string(data), -1)

	var images []string
	for _, substr := range substrings {
		images = append(images, substr[0])
	}

	return images, nil
}
