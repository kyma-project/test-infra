package updater

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/github/fakegithub"
)

func TestEnsurePRWithLabels(t *testing.T) {
	testCases := []struct {
		name   string
		client *fakegithub.FakeClient
	}{
		{
			name:   "pr is created",
			client: fakegithub.NewFakeClient(),
		},
		{
			name: "pr is updated",
			client: &fakegithub.FakeClient{
				PullRequests: map[int]*github.PullRequest{
					22: {Number: 22, User: github.User{Login: "k8s-ci-robot"}},
				},
				Issues: map[int]*github.Issue{
					22: {Number: 22},
				},
			},
		},
		{
			name: "existing labels are considered",
			client: &fakegithub.FakeClient{
				PullRequests: map[int]*github.PullRequest{
					42: {Number: 42, User: github.User{Login: "k8s-ci-robot"}},
				},
				Issues: map[int]*github.Issue{
					42: {
						Number: 42,
						Labels: []github.Label{{Name: "a"}},
					},
				},
				IssueLabelsAdded: []string{"org/repo#42:a"},
			},
		},
	}

	org, repo, labels := "org", "repo", []string{"a", "b"}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prNumberPtr, err := EnsurePRWithLabels(org, repo, "title", "body", "source", "branch", "matchTitle", PreventMods, tc.client, labels)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if n := len(tc.client.PullRequests); n != 1 {
				t.Fatalf("expected to find one PR, got %d", n)
			}

			expectedLabels := sets.NewString()
			for _, label := range labels {
				expectedLabels.Insert(fmt.Sprintf("%s/%s#%d:%s", org, repo, *prNumberPtr, label))
			}

			if diff := sets.NewString(tc.client.IssueLabelsAdded...).Difference(expectedLabels); len(diff) != 0 {
				t.Errorf("found labels do not match expected, diff: %v", diff)
			}
		})
	}
}

func TestEnsurePRWithQueryTokens(t *testing.T) {
	testCases := []struct {
		name        string
		client      *fakegithub.FakeClient
		expectedPR  int
		expectedErr bool
	}{
		{
			name:        "create new PR if no match",
			client:      fakegithub.NewFakeClient(),
			expectedPR:  0,
			expectedErr: false,
		},
		{
			name: "update existing PR",
			client: &fakegithub.FakeClient{
				PullRequests: map[int]*github.PullRequest{
					1: {Number: 1, Title: "old title", Body: "old body", User: github.User{Login: "k8s-ci-robot"}},
				},
				Issues: map[int]*github.Issue{
					1: {Number: 1},
				},
			},
			expectedPR:  1,
			expectedErr: false,
		},
	}

	org, repo, title, body, source, baseBranch := "org", "repo", "title", "body", "source", "baseBranch"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prNumber, err := EnsurePRWithQueryTokens(org, repo, title, body, source, baseBranch, "queryTokensString", AllowMods, tc.client)
			if (err != nil) != tc.expectedErr {
				t.Fatalf("error: %v, expected error: %v", err, tc.expectedErr)
			}
			if prNumber == nil {
				t.Fatalf("prNumber is nil")
			}
			if *prNumber != tc.expectedPR {
				t.Fatalf("expected PR number: %d, got: %d", tc.expectedPR, *prNumber)
			}
			// Dodatkowe logowanie
			fmt.Printf("PR number: %d\n", *prNumber)
			for number, pr := range tc.client.PullRequests {
				fmt.Printf("Existing PR in client - Number: %d, Title: %s\n", number, pr.Title)
			}
		})
	}
}

func TestUpdatePRWithQueryTokens(t *testing.T) {
	testCases := []struct {
		name        string
		client      *fakegithub.FakeClient
		expectedPR  *int
		expectedErr bool
	}{
		{
			name:        "no existing PRs",
			client:      fakegithub.NewFakeClient(),
			expectedPR:  nil,
			expectedErr: false,
		},
		{
			name: "update existing PR",
			client: &fakegithub.FakeClient{
				PullRequests: map[int]*github.PullRequest{
					1: {Number: 1, Title: "old title", Body: "old body", User: github.User{Login: "k8s-ci-robot"}},
				},
				Issues: map[int]*github.Issue{
					1: {Number: 1},
				},
			},
			expectedPR:  intPtr(1),
			expectedErr: false,
		},
	}

	org, repo, title, body, queryTokensString := "org", "repo", "title", "body", "queryTokensString"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prNumber, err := updatePRWithQueryTokens(org, repo, title, body, queryTokensString, tc.client)
			if (err != nil) != tc.expectedErr {
				t.Fatalf("error: %v, expected error: %v", err, tc.expectedErr)
			}
			if prNumber != tc.expectedPR && (prNumber == nil || tc.expectedPR == nil || *prNumber != *tc.expectedPR) {
				t.Fatalf("expected PR number: %v, got: %v", tc.expectedPR, prNumber)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
