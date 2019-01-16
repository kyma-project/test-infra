package release

import (
	"context"
	"os"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// GithubAPI exposes functions to interact with Github releases
type GithubAPI interface {
	CreateGithubRelease(ctx context.Context, releaseVersion string, releaseBody string, targetCommit string, isPreRelease bool) (*github.RepositoryRelease, *github.Response, error)
	UploadFile(ctx context.Context, releaseID int64, artifactName string, artifactFile *os.File) (*github.ReleaseAsset, *github.Response, error)
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
func (gap *githubAPIWrapper) CreateGithubRelease(ctx context.Context, releaseVersion string, releaseBody string, targetCommit string, isPreRelease bool) (*github.RepositoryRelease, *github.Response, error) {
	common.Shout("Creating release %s in %s/%s repository", releaseVersion, gap.repoOwner, gap.repoName)

	input := &github.RepositoryRelease{
		TagName:         &releaseVersion,
		TargetCommitish: &targetCommit,
		Name:            &releaseVersion,
		Body:            &releaseBody,
		Prerelease:      &isPreRelease,
	}

	return gap.githubClient.Repositories.CreateRelease(ctx, gap.repoOwner, gap.repoName, input)
}

// UploadFile adds file to a Github release
func (gap *githubAPIWrapper) UploadFile(ctx context.Context, releaseID int64, artifactName string, artifactFile *os.File) (*github.ReleaseAsset, *github.Response, error) {
	common.Shout("Uploading %s artifact", artifactName)

	opt := &github.UploadOptions{Name: artifactName}
	return gap.githubClient.Repositories.UploadReleaseAsset(ctx, gap.repoOwner, gap.repoName, releaseID, opt, artifactFile)
}
