package missing

import (
	"github.com/kyma-project/test-infra/pkg/image-url-helper/images"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createComponentImage(image images.Image) images.ComponentImage {
	componentImage := images.ComponentImage{
		Components: make(map[string]bool),
		Image:      image,
	}
	return componentImage
}

func errorNotNil(err error) bool {
	return err != nil
}

func TestImageMissing(t *testing.T) {
	tests := []struct {
		name     string
		image    images.ComponentImage
		expected bool
	}{
		{
			name: "Existing image with tag",
			image: createComponentImage(images.Image{
				ContainerRegistryURL:    "europe-docker.pkg.dev/kyma-project",
				ContainerRepositoryPath: "prod",
				Name:                    "automated-approver",
				Version:                 "v20250213-8adb8ce9",
				SHA:                     "",
			}),
			expected: false,
		},
		{
			name: "Existing image with SHA",
			image: createComponentImage(images.Image{
				ContainerRegistryURL:    "europe-docker.pkg.dev/kyma-project",
				ContainerRepositoryPath: "prod",
				Name:                    "automated-approver",
				Version:                 "",
				SHA:                     "1109d8e8187bcf502f0e950af18708030b2ef908907fd33b8b4cd485edcda077",
			}),
			expected: false,
		},
		{
			name: "Nonexistent image",
			image: createComponentImage(images.Image{
				ContainerRegistryURL:    "europe-docker.pkg.dev/kyma-project",
				ContainerRepositoryPath: "prod",
				Name:                    "automated-approver",
				Version:                 "missing-image",
				SHA:                     "",
			}),
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imageReference, err := parseImageReference(test.image)
			if err != nil {
				t.Errorf("failed to check for parse image reference: %s", err)
			}
			err = getImageError(imageReference)
			actual := errorNotNil(err)
			assert.New(t).Equal(test.expected, actual)
		})
	}
}
