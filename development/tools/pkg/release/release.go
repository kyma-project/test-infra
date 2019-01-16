package release

import (
	"context"
	"os"

	"github.com/kyma-project/test-infra/development/tools/pkg/file"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

const tmpFilePattern = "release-artifact-"

// Creator exposes a function to create complete Github releases
type Creator interface {
	CreateNewRelease(ctx context.Context, releaseVersion string, targetCommit string, releaseChangelogName string, localArtifactName string, clusterArtifactName string, isPreRelease bool) error
}

// creatorImpl provides functions to create a complete Github Release
type creatorImpl struct {
	github  GithubAPI
	storage StorageAPI
}

// NewCreator returns new instance of Creator interface
func NewCreator(github GithubAPI, storage StorageAPI) Creator {
	return &creatorImpl{
		github:  github,
		storage: storage,
	}
}

//CreateNewRelease .
func (c *creatorImpl) CreateNewRelease(ctx context.Context, releaseVersion string, targetCommit string, releaseChangelogName string, localArtifactName string, clusterArtifactName string, isPreRelease bool) error {

	//Changelog
	releaseChangelogData, err := c.storage.ReadBucketObject(ctx, releaseChangelogName)
	if err != nil {
		return errors.Wrapf(err, "while reading %s file", releaseChangelogName)
	}

	//Release
	release, _, err := c.github.CreateGithubRelease(ctx, releaseVersion, string(releaseChangelogData), targetCommit, isPreRelease)
	if err != nil {
		return errors.Wrap(err, "while creating Github release")
	}

	//LocalArtifactFile
	if err = c.createReleaseArtifact(ctx, *release.ID, localArtifactName); err != nil {
		return errors.Wrapf(err, "while creating release artifact: %s", localArtifactName)
	}

	//ClusterArtifactFile
	if err = c.createReleaseArtifact(ctx, *release.ID, clusterArtifactName); err != nil {
		return errors.Wrapf(err, "while creating release artifact: %s", clusterArtifactName)
	}

	return nil
}

func (c *creatorImpl) createReleaseArtifact(ctx context.Context, releaseID int64, artifactName string) error {

	artifactData, err := c.storage.ReadBucketObject(ctx, artifactName)
	if err != nil {
		return errors.Wrapf(err, "while reading %s file", artifactName)
	}

	artifactFile, err := file.SaveDataToTmpFile(artifactData, tmpFilePattern)
	if err != nil {
		return errors.Wrap(err, "while saving temporary file")
	}

	defer os.Remove(artifactFile.Name())

	_, _, err = c.uploadFile(ctx, releaseID, artifactFile, artifactName)
	if err != nil {
		return errors.Wrapf(err, "while uploading %s file", artifactName)
	}

	return nil

}

func (c *creatorImpl) uploadFile(ctx context.Context, releaseID int64, file *os.File, fileName string) (*github.ReleaseAsset, *github.Response, error) {

	artifactFile, err := os.Open(file.Name())
	if err != nil {
		return nil, nil, err
	}

	defer artifactFile.Close()

	return c.github.UploadFile(ctx, releaseID, fileName, artifactFile)

}
