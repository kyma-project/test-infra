package testuntrusted

import (
	"fmt"

	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/plugins"
)


func handlePR(pr github.PullRequestEvent) error {
	org, repo, a := orgRepoAuthor(pr.PullRequest)
	author := string(a)
	num := pr.PullRequest.Number

	baseSHA := ""
	baseSHAGetter := func() (string, error) {
		var err error
		baseSHA, err = c.GitHubClient.GetRef(org, repo, "heads/"+pr.PullRequest.Base.Ref)
		if err != nil {
			return "", fmt.Errorf("failed to get baseSHA: %w", err)
		}
		return baseSHA, nil
	}
	headSHAGetter := func() (string, error) {
		return pr.PullRequest.Head.SHA, nil
	}

	presubmits := getPresubmits(c.Logger, c.GitClient, c.Config, org+"/"+repo, baseSHAGetter, headSHAGetter)
	if len(presubmits) == 0 {
		return nil
	}

	if baseSHA == "" {
		if _, err := baseSHAGetter(); err != nil {
			return err
		}
	}

	switch pr.Action {
	case github.PullRequestActionOpened:
