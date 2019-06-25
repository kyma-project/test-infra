package release

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// GithubAPI exposes functions to interact with Github releases
type GithubAPI interface {
	CreateGithubRelease(ctx context.Context, opts *Options) (*github.RepositoryRelease, *github.Response, error)
	UploadContent(ctx context.Context, releaseID int64, artifactName string, reader io.Reader, size int64) (*github.Response, error)
}

// githubAPIWrapper implements functions to interact with Github releases
type githubAPIWrapper struct {
	githubClient *github.Client
	repoOwner    string
	repoName     string
}

// NewGithubAPI returns implementation of githubAPI
func NewGithubAPI(ctx context.Context, githubAccessToken, repoOwner, repoName string) GithubAPI {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)

	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)

	return &githubAPIWrapper{
		githubClient: githubClient,
		repoOwner:    repoOwner,
		repoName:     repoName,
	}
}

// CreateGithubRelease creates a Github release
func (gaw *githubAPIWrapper) CreateGithubRelease(ctx context.Context, opts *Options) (*github.RepositoryRelease, *github.Response, error) {
	common.Shout("Creating release %s in %s/%s repository", opts.Version, gaw.repoOwner, gaw.repoName)

	input := &github.RepositoryRelease{
		TagName:         &opts.Version,
		TargetCommitish: &opts.TargetCommit,
		Name:            &opts.Version,
		Body:            &opts.Body,
		Prerelease:      &opts.IsPreRelease,
	}

	return gaw.githubClient.Repositories.CreateRelease(ctx, gaw.repoOwner, gaw.repoName, input)
}

// UploadContent creates an asset by uploading a file into a release repository
func (gaw *githubAPIWrapper) UploadContent(ctx context.Context, releaseID int64, artifactName string, reader io.Reader, size int64) (*github.Response, error) {

	common.Shout("Uploading %s artifact", artifactName)

	u := fmt.Sprintf("repos/%s/%s/releases/%d/assets", gaw.repoOwner, gaw.repoName, releaseID)

	opt := &github.UploadOptions{Name: artifactName}

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, err
	}

	req, err := gaw.githubClient.NewUploadRequest(u, reader, size, "")
	if err != nil {
		return nil, err
	}

	asset := new(github.ReleaseAsset)

	return gaw.githubClient.Do(ctx, req, asset)
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}
