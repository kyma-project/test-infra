package release

import (
	"context"
	"log"

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

//CreateNewRelease .
func (c *creatorImpl) CreateNewRelease(ctx context.Context, relOpts *Options, artifactNames ...string) error {

	//Release
	release, _, err := c.github.CreateGithubRelease(ctx, relOpts)
	if err != nil {
		return errors.Wrap(err, "while creating Github release")
	}

	log.Printf("Successfully created release: %v", *release.ID)

	return nil
}
