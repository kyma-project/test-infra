package release

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

// Creator exposes a function to create complete Github releases
type Creator interface {
	CreateNewRelease(ctx context.Context, relOpts *Options) error
}

// creatorImpl provides functions to create a complete Github Release
type creatorImpl struct {
	github GithubAPI
}

// NewCreator returns implementation of Creator interface
func NewCreator(github GithubAPI) Creator {
	return &creatorImpl{
		github: github,
	}
}

// CreateNewRelease .
func (c *creatorImpl) CreateNewRelease(ctx context.Context, relOpts *Options) error {

	//Release
	release, _, err := c.github.CreateGithubRelease(ctx, relOpts)
	if err != nil {
		return errors.Wrap(err, "while creating Github release")
	}

	if err = c.createReleaseArtifact(ctx, *release.ID, relOpts.KymaComponentsName, relOpts.KymaComponentsPath); err != nil {
		return errors.Wrapf(err, "while creating release artifact: %s", relOpts.KymaComponentsName)
	}

	return nil
}

func (c *creatorImpl) createReleaseArtifact(ctx context.Context, releaseID int64, artifactName, componentsPath string) error {

	components, err := os.Open(componentsPath)
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
