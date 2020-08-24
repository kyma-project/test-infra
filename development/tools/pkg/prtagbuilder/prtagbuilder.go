package prtagbuilder

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/google/go-github/v31/github"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

const (
	ghProxyURL = "http://ghproxy/"
)

var (
	client *github.Client
	ctx    = context.Background()
)

// findPRNumber match commit message with regex to extract pull request number. By default github add pr number to the commit message.
func findPRNumber(commit *github.RepositoryCommit) string {
	re := regexp.MustCompile(`^.*\(#(?P<prNumber>\d*)\)\s*$`)
	matches := re.FindStringSubmatch(*commit.Commit.Message)
	if len(matches) != 2 {
		logrus.Fatalf("failed find PR number in commit message, found %s matched strings", len(matches))
	}
	return matches[1]
}

// verifyPR checks if pull request merge commit match provided commit SHA.
func verifyPR(pr *github.PullRequest, commitSHA string) bool {
	if *pr.Merged {
		if *pr.MergeCommitSHA != commitSHA {
			logrus.Fatalf("commit SHA and matched PR merge commit SHA doesn't match")
		}
	}
	return true
}

func BuildPrTag() {
	// check if prtagbuilder was called on presubmit, fail if true
	if _, present := os.LookupEnv("PULL_NUMBER"); present {
		logrus.Fatalf("prtagbuilder was called on presubmit, failing")
	}
	// get git base reference from postsubmit environment variables
	jobSpec, err := downwardapi.ResolveSpecFromEnv()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to read JOB_SPEC prowjob env")
	}
	// create github client with ghproxy URL
	client, err = github.NewEnterpriseClient(ghProxyURL, ghProxyURL, nil)
	if err != nil {
		logrus.WithError(err).Fatalf("failed get new github client")
	}
	// get api meta data through ghproxy
	_, _, err = client.APIMeta(ctx)
	if err != nil {
		logrus.WithError(err).Warnf("failed connecting to ghproxy")
		// fallback to create github client with default public github URL because ghproxy didn't respond
		client = github.NewClient(nil)
	}
	// get commit details for base sha
	commit, _, err := client.Repositories.GetCommit(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, jobSpec.Refs.BaseSHA)
	if err != nil {
		logrus.WithError(err).Fatalf("failed get commit %s", jobSpec.Refs.BaseSHA)
	}
	// extract pull request number from commit message
	prNumber, err := strconv.Atoi(findPRNumber(commit))
	if err != nil {
		logrus.WithError(err).Fatalf("failed convert PR number to integer")
	}
	// get pull request details for extracted pr
	pr, _, err := client.PullRequests.Get(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, prNumber)
	if err != nil {
		logrus.WithError(err).Fatalf("failed get Pull Request number %s", prNumber)
	}
	// check if correct pr was found
	if verifyPR(pr, jobSpec.Refs.BaseSHA) {
		// print PR image tag
		fmt.Printf("PR-%d", prNumber)
	}
}
