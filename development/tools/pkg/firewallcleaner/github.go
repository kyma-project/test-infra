package firewallcleaner

import (
	"context"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// GithubAPI exposes functions to interact with Github releases
type GithubAPI interface {
	ClosedPullRequests(ctx context.Context) []*github.PullRequest
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

//ClosedPullRequests returns a list of pull requests, that are closed for a long enough time (currently 3 hour window).
func (g *githubAPIWrapper) ClosedPullRequests(ctx context.Context) []*github.PullRequest {
	opts := &github.PullRequestListOptions{State: "closed"}
	pulls, resp, err := g.githubClient.PullRequests.List(ctx, g.repoOwner, g.repoName, opts)
	var prList []*github.PullRequest
	if err != nil {
		common.Shout("error reading PRs from '%s'", g.repoName)
	}
	if resp.StatusCode != 200 {
		common.Shout("something does not add up to 200: %s", resp.Status)
	}
	for _, p := range pulls {
		common.Shout("\"%s (PR #%d)\" is %s", p.GetTitle(), p.GetNumber(), p.GetState())
		now := time.Now()
		cutoffTime := now.Add(time.Hour * 3 * -1).Add(time.Minute * 15 * -1) // 3 hours 15 minutes to be definitely after the 3 hour lifetime window
		if "closed" == p.GetState() && p.GetClosedAt().Before(cutoffTime) {
			prList = append(prList, p)
		}
	}
	return prList
}
