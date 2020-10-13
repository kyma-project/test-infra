package prtagbuilder

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v31/github"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

var (
	client *github.Client
	ctx    = context.Background()
)

// findPRNumber match commit message with regex to extract pull request number. By default github add pr number to the commit message.
func findPRNumber(commit *github.RepositoryCommit) string {
	re := regexp.MustCompile(`(?s)^.*\(#(?P<prNumber>\d*)?\)\s*.*$`)
	messageReader := strings.NewReader(commit.Commit.GetMessage())
	scanner := bufio.NewScanner(messageReader)
	scanner.Scan()
	matches := re.FindStringSubmatch(scanner.Text())
	if len(matches) != 2 {
		logrus.Fatalf("failed find PR number in first line of commit message: %s, found %d matched strings", scanner.Text(), len(matches))
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

// BuildPrTag will extract PR number from commit message, search PR, validate if correct PR was found and print pr tag.
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
	// create github client
	client = github.NewClient(nil)

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
		logrus.WithError(err).Fatalf("failed get Pull Request number %d", prNumber)
	}
	// check if correct pr was found
	if verifyPR(pr, jobSpec.Refs.BaseSHA) {
		// print PR image tag
		fmt.Printf("PR-%d", prNumber)
	}
}
