package prtagbuilder

import (
	"context"
	"fmt"
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

func findPRNumber(commit *github.RepositoryCommit) string {
	re := regexp.MustCompile(`^.*\(#(?P<prNumber>\d*)\)\s*$`)
	matches := re.FindStringSubmatch(*commit.Commit.Message)
	if len(matches) != 2 {
		logrus.Fatalf("failed find PR number in commit message, found %s matched strings", len(matches))
	}
	return matches[1]
}

func verifyPR(pr *github.PullRequest, commitSHA string) bool {
	if *pr.Merged {
		if *pr.MergeCommitSHA != commitSHA {
			logrus.Fatalf("commit SHA and matched PR merge commit SHA doesn't match")
		}
	}
	return true
}
func BuildPrTag() {
	jobSpec, err := downwardapi.ResolveSpecFromEnv()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to read JOB_SPEC prowjob env")
	}
	client, err = github.NewEnterpriseClient(ghProxyURL, ghProxyURL, nil)
	if err != nil {
		logrus.WithError(err).Fatalf("failed get new github client")
	}
	_, _, err = client.APIMeta(ctx)
	if err != nil {
		logrus.WithError(err).Warnf("failed connecting to ghproxy")
		client = github.NewClient(nil)
	}
	commit, _, err := client.Repositories.GetCommit(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, jobSpec.Refs.BaseSHA)
	if err != nil {
		logrus.WithError(err).Fatalf("failed get commit %s", jobSpec.Refs.BaseSHA)
	}
	prNumber, err := strconv.Atoi(findPRNumber(commit))
	if err != nil {
		logrus.WithError(err).Fatalf("failed convert PR number to intiger")
	}
	pr, _, err := client.PullRequests.Get(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, prNumber)
	if err != nil {
		logrus.WithError(err).Fatalf("failed get Pull Request number %s", prNumber)
	}
	if verifyPR(pr, jobSpec.Refs.BaseSHA) {
		fmt.Printf("PR%d", prNumber)
	}
}
