package release

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

// Creator exposes a function to create complete Github releases
type Creator interface {
	CreateNewRelease(ctx context.Context, relOpts *Options, artifactNames ...string) error
}

// creatorImpl provides functions to create a complete Github Release
type creatorImpl struct {
	github  GithubAPI
	storage StorageAPI
}

// NewCreator returns implementation of Creator interface
func NewCreator(github GithubAPI, storage StorageAPI) Creator {
	return &creatorImpl{
		github:  github,
		storage: storage,
	}
}

// CreateNewRelease .
func (c *creatorImpl) CreateNewRelease(ctx context.Context, relOpts *Options, artifactNames ...string) error {

	//Release
	release, _, err := c.github.CreateGithubRelease(ctx, relOpts)
	if err != nil {
		return errors.Wrap(err, "while creating Github release")
	}

	for _, artifact := range artifactNames {
		if err = c.createReleaseArtifact(ctx, *release.ID, artifact); err != nil {
			return errors.Wrapf(err, "while creating release artifact: %s", artifact)
		}
	}

	return nil
}

func (c *creatorImpl) createReleaseArtifact(ctx context.Context, releaseID int64, artifactName string) error {

	components, err := os.Open("installation/resources/components.yaml")
	if err != nil {
		return errors.Wrapf(err, "while opening components.yaml file")
	}
	defer components.Close()
	fi, err := components.Stat()
	if err != nil {
		return errors.Wrapf(err, "while getting components.yaml file info")
	}

	_, err = c.github.UploadContent(ctx, releaseID, artifactName, components, fi.Size())
	if err != nil {
		return errors.Wrapf(err, "while uploading %s file", artifactName)
	}

	return nil

}
