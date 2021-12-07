package main

import (
	"k8s.io/test-infra/prow/github"
	"testing"
)

type fakeGhClient struct {
}

func GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error) {
	prc := []github.PullRequestChange{
		{
			Filename: "README.md",
		},
		{
			Filename: "some_other_file.md",
		},
	}
	return prc, nil
}

func Test_Server(t *testing.T) {
	_ = Plugin{}
}
