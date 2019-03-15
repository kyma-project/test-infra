package firewallcleaner

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// DUPLICATE CODE together with github.go in release folder, consolidate

// GithubAPI exposes functions to interact with Github releases
type GithubAPI interface {
	OpenPullRequests(ctx context.Context) []*github.PullRequest
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

func (g *githubAPIWrapper) OpenPullRequests(ctx context.Context) []*github.PullRequest {
	opts := &github.PullRequestListOptions{State: "all"}
	pulls, resp, err := g.githubClient.PullRequests.List(ctx, g.repoOwner, g.repoName, opts)
	if err != nil {
		common.Shout("error reading PRs from '%s'", g.repoName)
	}
	if resp.StatusCode != 200 {
		common.Shout("something does not add up to 200: %s", resp.Status)
	}
	for _, p := range pulls {
		common.Shout("PR #%d is open", p.GetID())
	}
	return pulls
}
