package release

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// GithubAPI exposes functions to interact with Github releases
type GithubAPI interface {
	CreateGithubRelease(ctx context.Context, opts *Options) (*github.RepositoryRelease, *github.Response, error)
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
