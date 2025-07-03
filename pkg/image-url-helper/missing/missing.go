package missing

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/kyma-project/test-infra/pkg/image-url-helper/image"
)

// CheckForMissingImages checks if images exist and returns ComponentImageMap of nonexistent images
func CheckForMissingImages(allImages image.ComponentImageMap, missingImages image.ComponentImageMap) error {

	for imageURL, image := range allImages {
		imageReference, err := parseImageReference(image)
		if err != nil {
			return err
		}

		err = getImageError(imageReference)
		if err != nil {
			// failed to fetch, add to list of non-existent images
			missingImages[imageURL] = image
		}
	}

	return nil
}

func parseImageReference(image image.ComponentImage) (name.Reference, error) {
	return name.ParseReference(image.Image.FullImageURL())
}

// getImageError checks if particular image exists
func getImageError(imageReference name.Reference) error {
	_, err := remote.Image(imageReference)
	return err
}
