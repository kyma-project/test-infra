package missing

import (
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/common"
)

// CheckForMissingImages checks if images exist and returns ComponentImageMap of nonexistent images
func CheckForMissingImages(allImages common.ComponentImageMap, missingImages common.ComponentImageMap) error {

	for imageURL, image := range allImages {
		imageReference, err := name.ParseReference(image.Image.FullImageURL())
		if err != nil {
			return err
		}
		_, err = remote.Image(imageReference)
		if err != nil {
			if !strings.Contains(err.Error(), "Failed to fetch") {
				// unknown error, fail here
				return err
			}

			// failed to fetch, add to list of non-existent images
			componentNames := make([]string, 0)
			for component := range image.Components {
				componentNames = append(componentNames, component)
			}

			missingImages[imageURL] = image
		}
	}

	return nil
}
