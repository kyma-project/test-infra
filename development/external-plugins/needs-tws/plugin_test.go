package main

import (
	"k8s.io/test-infra/prow/github"
	"testing"
)

type fakeGhClient struct {
	githubClient
}

func (f fakeGhClient) GetIssueLabels(org, repo string, number int) ([]github.Label, error) {
	il := []github.Label{
		{Name: DefaultNeedsTwsLabel},
		{Name: "foo-bar"},
	}
	return il, nil
}

func (f fakeGhClient) GetSingleCommit(org, repo, SHA string) (github.RepositoryCommit, error) {
	rc := github.RepositoryCommit{
		Files: []github.CommitFile{
			{
				Filename: "README.md",
			},
		},
	}
	return rc, nil
}

func (f fakeGhClient) AddLabel(org, repo string, number int, label string) error {
	return nil
}
func (f fakeGhClient) RemoveLabel(org, repo string, number int, label string) error {
	return nil
}

func Test_handlePullRequest(t *testing.T) {
	_ = Plugin{
		tokenGenerator: nil,
		ghc:            fakeGhClient{},
	}
}
