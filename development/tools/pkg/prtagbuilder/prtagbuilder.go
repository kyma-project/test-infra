package prtagbuilder

import (
	"fmt"
	"net/http"

	"k8s.io/test-infra/prow/config/secret"

	"github.com/sirupsen/logrus"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

type githubClient interface {
	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
	GetSingleCommit(org, repo, SHA string) (github.SingleCommit, error)
}

type SingleCommit struct {
	Commit struct {
		Tree struct {
			SHA string `json:"sha"`
		} `json:"tree"`
	} `json:"commit"`
}

// GetSingleCommit returns a single commit.
//
// See https://developer.github.com/v3/repos/#get
func GetSingleCommit(c github.Client, org, repo, SHA string) (SingleCommit, error) {
	durationLogger := c.log("GetSingleCommit", org, repo, SHA)
	defer durationLogger()

	var commit SingleCommit
	_, err := c.request(&request{
		method:    http.MethodGet,
		path:      fmt.Sprintf("/repos/%s/%s/commits/%s", org, repo, SHA),
		exitCodes: []int{200},
	}, &commit)
	return commit, err
}

func BuildPrTag(ghOptions prowflagutil.GitHubOptions) {
	var secretAgent *secret.Agent
	if ghOptions.TokenPath != "" {
		secretAgent = &secret.Agent{}
		if err := secretAgent.Start([]string{ghOptions.TokenPath}); err != nil {
			logrus.WithError(err).Fatal("Failed to start secret agent")
		}
	}
	jobSpec, err := downwardapi.ResolveSpecFromEnv()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to read JOB_SPEC prowjob env")
	}
	var ghClient githubClient
	ghClient, err = ghOptions.GitHubClient(secretAgent, false)
	if err != nil {
		logrus.WithError(err).Fatalf("failed create github client")
	}
	var commit github.SingleCommit
	commit, err = ghClient.GetSingleCommit(jobSpec.Refs.Org, jobSpec.Refs.Repo, jobSpec.Refs.BaseSHA)
	if err != nil {
		logrus.WithError(err).Fatal("failed get commit")
	}
	fmt.Printf("%v", commit)
}
