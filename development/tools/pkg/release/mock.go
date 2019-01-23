package release

import (
	"context"
	"io"
	"io/ioutil"
	"strings"

	"github.com/google/go-github/github"
)

// FakeKymaVersionReader is a fake kymaVersionReader for test purposes
type FakeKymaVersionReader struct{}

// Read is a fake implementation of a Read method
func (fkvr *FakeKymaVersionReader) Read(filePath string) (string, bool, error) {
	return filePath, strings.Contains(filePath, "rc"), nil
}

// FakeStorageAPIWrapper is a fake storageAPIWrapper for test purposes
type FakeStorageAPIWrapper struct {
	TimesReadBucketObjectCalled int
}

// ReadBucketObject is a fake implementation of ReadBucketObject func
func (fsaw *FakeStorageAPIWrapper) ReadBucketObject(ctx context.Context, fileName string) (io.ReadCloser, int64, error) {
	fsaw.TimesReadBucketObjectCalled++
	return ioutil.NopCloser(strings.NewReader("test artifact data for " + fileName)), 100, nil

}

// FakeGithubAPIWrapper is a fake githubAPIWrapper for test purposes
type FakeGithubAPIWrapper struct {
	Release               *github.RepositoryRelease
	Assets                []*github.ReleaseAsset
	AssetCount            int
	TimesUploadFileCalled int
}

// CreateGithubRelease is a fake implementation of CreateGithubRelease func
func (fgaw *FakeGithubAPIWrapper) CreateGithubRelease(ctx context.Context, opts *Options) (*github.RepositoryRelease, *github.Response, error) {

	fakeID := int64(1)

	input := &github.RepositoryRelease{
		ID:              &fakeID,
		TagName:         &opts.Version,
		TargetCommitish: &opts.TargetCommit,
		Name:            &opts.Version,
		Body:            &opts.Body,
		Prerelease:      &opts.IsPreRelease,
	}

	fgaw.Release = input

	return input, nil, nil
}

// UploadContent is a fake implementation of UploadContent func
func (fgaw *FakeGithubAPIWrapper) UploadContent(ctx context.Context, releaseID int64, artifactName string, reader io.Reader, size int64) (*github.Response, error) {

	currID := int64(fgaw.AssetCount)

	asset := &github.ReleaseAsset{
		ID:   &currID,
		Name: &artifactName,
	}

	fgaw.AssetCount++
	fgaw.TimesUploadFileCalled++

	fgaw.Assets = append(fgaw.Assets, asset)

	return nil, nil
}
