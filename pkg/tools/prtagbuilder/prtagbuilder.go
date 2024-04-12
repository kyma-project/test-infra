package prtagbuilder

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v48/github"
	"sigs.k8s.io/prow/prow/pod-utils/downwardapi"
)

type githubRepoService interface {
	GetBranch(ctx context.Context, owner, repo, branch string, followRedirects bool) (*github.Branch, *github.Response, error)
	GetCommit(ctx context.Context, owner, repo, sha string, opts *github.ListOptions) (*github.RepositoryCommit, *github.Response, error)
	// other repository methods used in your app
}

type githubPullRequestsService interface {
	Get(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error)
	// other pull requests methods used in your app
}

// GitHubClient is an implementation of go-github client with needed services only.
type GitHubClient struct {
	Repositories githubRepoService
	PullRequests githubPullRequestsService
	// optionally store and export the underlying *github.Client
	// if you want easy access to client.Rate or other fields
}

// NewGitHubClient returns new instance of go-github GitHubClient implementation.
func NewGitHubClient(httpClient *http.Client) *GitHubClient {
	client := github.NewClient(httpClient)
	// optionally set client.BaseURL, client.UserAgent, etc

	return &GitHubClient{
		Repositories: client.Repositories,
		PullRequests: client.PullRequests,
		// any other services your app uses
	}
}

// findPRNumber match commit message with regex to extract pull request number. By default github add pr number to the commit message.
func findPRNumber(commit *github.RepositoryCommit) (string, error) {
	re := regexp.MustCompile(`(?s)^.*\(#(?P<prNumber>\d*)?\)\s*$`)
	messageReader := strings.NewReader(commit.Commit.GetMessage())
	scanner := bufio.NewScanner(messageReader)
	scanner.Scan()
	matches := re.FindStringSubmatch(scanner.Text())
	if len(matches) != 2 {
		return "", fmt.Errorf("failed find PR number in first line of commit message: %s, found %d matched strings", scanner.Text(), len(matches))
	}
	return matches[1], nil
}

// verifyPR checks if pull request merge commit match provided commit SHA.
func verifyPR(pr *github.PullRequest, commitSHA string) error {
	if *pr.Merged {
		if *pr.MergeCommitSHA != commitSHA {
			return fmt.Errorf("commit SHA %s and matched PR merge commit SHA %s, doesn't match", commitSHA, *pr.MergeCommitSHA)
		}
	}
	return nil
}

// BuildPrTag will extract PR number from commit message, search PR, validate if correct PR was found and print pr tag.
func BuildPrTag(jobSpec *downwardapi.JobSpec, fromFlags bool, numberOnly bool, client *GitHubClient) (string, error) {
	var (
		ctx    = context.Background()
		commit *github.RepositoryCommit
	)

	if fromFlags {
		// get commit for a branch
		branch, _, err := client.Repositories.GetBranch(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, jobSpec.Refs.BaseRef, true)
		if err != nil {
			return "", fmt.Errorf("failed get branch %s, got error: %w", jobSpec.Refs.BaseRef, err)
		}
		commit = branch.GetCommit()
		jobSpec.Refs.BaseSHA = *commit.SHA
	} else {
		var err error
		// get git base reference from postsubmit environment variables
		jobSpec, err = downwardapi.ResolveSpecFromEnv()
		if err != nil {
			return "", fmt.Errorf("failed to read JOB_SPEC env, got error: %w", err)
		}
		// get commit details for base sha
		commit, _, err = client.Repositories.GetCommit(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, jobSpec.Refs.BaseSHA, nil)
		if err != nil {
			return "", fmt.Errorf("failed get commit %s, got error: %w", jobSpec.Refs.BaseSHA, err)
		}
	}
	// extract pull request number from commit message
	prNumberAsString, err := findPRNumber(commit)
	if err != nil {
		return "", fmt.Errorf("failed get PR number, got error: %w", err)
	}
	prNumber, err := strconv.Atoi(prNumberAsString)
	if err != nil {
		return "", fmt.Errorf("failed convert PR number to integer, got error: %w", err)
	}
	// get pull request details for extracted pr
	pr, _, err := client.PullRequests.Get(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, prNumber)
	if err != nil {
		return "", fmt.Errorf("failed get Pull Request number %d, got error: %w", prNumber, err)
	}
	// check if correct pr was found
	if err := verifyPR(pr, jobSpec.Refs.BaseSHA); err != nil {
		return "", fmt.Errorf("pr verification failed, got error: %w", err)
	}
	if numberOnly {
		return fmt.Sprintf("%d", prNumber), nil
	}
	// print PR image tag
	return fmt.Sprintf("PR-%d", prNumber), nil
}
