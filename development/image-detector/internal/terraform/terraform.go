package terraform

import (
	"io"
	"os"
	"regexp"
)

func Extract(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return extract(f)
}

func extract(reader io.Reader) ([]string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile("image[ ]+=[ ]+\"(.*[[a-z/-]+:[a-z0-9.-]+?)\"")
	substrings := re.FindAllStringSubmatch(string(data), -1)

	var images []string
	for _, substr := range substrings {
		images = append(images, substr[1:]...)
	}

	return images, nil
}
