package missing

import (
	"github.com/kyma-project/test-infra/pkg/image-url-helper/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createComponentImage(image common.Image) common.ComponentImage {
	componentImage := common.ComponentImage{
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
		image    common.ComponentImage
		expected bool
	}{
		{
			name: "Existing image with tag",
			image: createComponentImage(common.Image{
				ContainerRegistryURL:    "eu.gcr.io/kyma-project",
				ContainerRepositoryPath: "tpi",
				Name:                    "k8s-tools",
				Version:                 "20220516-9f87ea89",
				SHA:                     "",
			}),
			expected: false,
		},
		{
			name: "Existing image with SHA",
			image: createComponentImage(common.Image{
				ContainerRegistryURL:    "eu.gcr.io/kyma-project",
				ContainerRepositoryPath: "tpi",
				Name:                    "k8s-tools",
				Version:                 "",
				SHA:                     "02317e1d351951f85b96bef7f058fc40181e3b93ac4109f3f4858a8e36ec0962",
			}),
			expected: false,
		},
		{
			name: "Nonexistent image",
			image: createComponentImage(common.Image{
				ContainerRegistryURL:    "eu.gcr.io/kyma-project",
				ContainerRepositoryPath: "tpi",
				Name:                    "k8s-tools",
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
