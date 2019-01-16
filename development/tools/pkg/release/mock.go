package release

import (
	"context"
	"os"

	"github.com/google/go-github/github"
)

// FakeStorageAPIWrapper is a fake storageAPIWrapper for test purposes
type FakeStorageAPIWrapper struct {
	TimesReadBucketObjectCalled int
}

// ReadBucketObject is a fake implementation of ReadBucketObject func
func (fsaw *FakeStorageAPIWrapper) ReadBucketObject(ctx context.Context, fileName string) ([]byte, error) {
	fsaw.TimesReadBucketObjectCalled++
	return []byte("test artifact data for " + fileName), nil
}

// FakeGithubAPIWrapper is a fake githubAPIWrapper for test purposes
type FakeGithubAPIWrapper struct {
	Release               *github.RepositoryRelease
	Assets                []*github.ReleaseAsset
	AssetCount            int
	TimesUploadFileCalled int
}

// CreateGithubRelease is a fake implementation of CreateGithubRelease func
func (fgaw *FakeGithubAPIWrapper) CreateGithubRelease(ctx context.Context, releaseVersion string, releaseBody string, targetCommit string, isPreRelease bool) (*github.RepositoryRelease, *github.Response, error) {

	fakeID := int64(1)

	input := &github.RepositoryRelease{
		ID:              &fakeID,
		TagName:         &releaseVersion,
		TargetCommitish: &targetCommit,
		Name:            &releaseVersion,
		Body:            &releaseBody,
		Prerelease:      &isPreRelease,
	}

	fgaw.Release = input

	return input, nil, nil
}

// UploadFile is a fake implementation of UploadFile func
func (fgaw *FakeGithubAPIWrapper) UploadFile(ctx context.Context, releaseID int64, artifactName string, artifactFile *os.File) (*github.ReleaseAsset, *github.Response, error) {

	currID := int64(fgaw.AssetCount)

	asset := &github.ReleaseAsset{
		ID:   &currID,
		Name: &artifactName,
	}

	fgaw.AssetCount++
	fgaw.TimesUploadFileCalled++

	fgaw.Assets = append(fgaw.Assets, asset)

	return nil, nil, nil

}
