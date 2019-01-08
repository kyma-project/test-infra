package githubrelease

import (
	"context"
	"os"

	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// GithubAPIWrapper abstracts Github API
type GithubAPIWrapper struct {
	Context   context.Context
	Client    *github.Client
	RepoOwner string
	RepoName  string
}

// CreateGithubRelease creates github release
func (gap *GithubAPIWrapper) CreateGithubRelease(releaseVersion string, releaseBody string, targetCommit string, isPreRelease bool) (*github.RepositoryRelease, *github.Response, error) {
	common.Shout("Creating release %s in %s/%s repository", releaseVersion, gap.RepoOwner, gap.RepoName)

	input := &github.RepositoryRelease{
		TagName:         &releaseVersion,
		TargetCommitish: &targetCommit,
		Name:            &releaseVersion,
		Body:            &releaseBody,
		Prerelease:      &isPreRelease,
	}

	return gap.Client.Repositories.CreateRelease(gap.Context, gap.RepoOwner, gap.RepoName, input)
}

// UploadArtifact adds file to a github release
func (gap *GithubAPIWrapper) UploadArtifact(releaseID int64, artifactName string, artifactFile *os.File) (*github.ReleaseAsset, *github.Response, error) {
	common.Shout("Uploading %s artifact", artifactName)

	opt := &github.UploadOptions{Name: artifactName}
	return gap.Client.Repositories.UploadReleaseAsset(gap.Context, gap.RepoOwner, gap.RepoName, releaseID, opt, artifactFile)
}
