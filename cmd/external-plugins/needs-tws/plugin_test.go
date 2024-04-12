package main

import (
	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	"testing"

	"go.uber.org/zap"
	"sigs.k8s.io/prow/prow/git/v2"
	"sigs.k8s.io/prow/prow/github"
	"sigs.k8s.io/prow/prow/github/fakegithub"
	"sigs.k8s.io/prow/prow/repoowners"
)

type fakeAliases struct {
	ownersAliases
	// This has to be normalized list of aliases
	// lowercase and without any GitHub prefixes
	Aliases repoowners.RepoAliases
}

type fakeClientFactory struct {
	git.ClientFactory
}

type fakeRepoClient struct {
	git.RepoClient
}

func (f fakeAliases) LoadOwnersAliases(l *zap.SugaredLogger, basedir, filename string) (repoowners.RepoAliases, error) {
	return f.Aliases, nil
}

func (f fakeRepoClient) Directory() string {
	return ""
}

func (f fakeClientFactory) ClientFor(org, repo string) (git.RepoClient, error) {
	return fakeRepoClient{}, nil
}

func Test_HandlePullRequest(t *testing.T) {
	SHA := "9448a2cb0a3915ac956685de8ffb3d4ef55fbc05"
	twsLabel := "org/repo#101:do-not-merge/missing-docs-review"
	testcases := []struct {
		name                string
		event               github.PullRequestEvent
		changes             []github.PullRequestChange
		IssueLabelsAdded    int
		IssueLabelsRemoved  int
		IssueLabelsExisting []string
		IssueCommentsAdded  int
		Reviews             []github.Review
	}{
		{
			name: "pr_opened, files changed, add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionOpened,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "README.md",
				},
			},
			IssueLabelsAdded: 1,
		},
		{
			name: "pr_opened, files not, changed, do not add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionOpened,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "path/to/file.go",
				},
			},
		},
		{
			name: "pr_opened, files not changed, do not add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionOpened,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Name: "org"},
				},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "path/to/other.file",
				},
				{
					Filename: "path/to/cmd/main.go",
				},
			},
		},
		{
			name: "pr_synchronize, files changed, add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionSynchronize,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "README.md",
				},
			},
			IssueLabelsAdded: 1,
		},
		{
			name: "pr_synchronize, files changed, already has a label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionSynchronize,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "README.md",
				},
			},
			IssueLabelsExisting: []string{twsLabel},
		},
		{
			name: "pr_synchronize, files not changed, do not add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionSynchronize,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
		},
		{
			name: "pr_synchronize, files not changed, label present, remove label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionSynchronize,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			IssueLabelsExisting: []string{twsLabel},
			IssueLabelsRemoved:  1,
		},
		{
			name: "pr_opened is a draft",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionOpened,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
					Draft: true,
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
		},
		{
			name: "pr_closed",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionClosed,
			},
		},
		{
			name: "pr_labeled",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionLabeled,
			},
		},
		{
			name: "pr_unlabeled by a bot",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionUnlabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
				Sender: github.User{Login: fakegithub.Bot},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "README.md",
				},
			},
		},
		{
			name: "pr_unlabeled not a documentation label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionUnlabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
				Sender: github.User{Login: "collaborator"},
				Label:  github.Label{Name: "not-a-docs-label"},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "README.md",
				},
			},
		},
		{
			name: "pr_unlabeled not by a collaborator",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionUnlabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
				Sender: github.User{Login: "not-a-collaborator"},
				Label:  github.Label{Name: DefaultNeedsTwsLabel},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "README.md",
				},
			},
			IssueCommentsAdded: 1,
			IssueLabelsAdded:   1,
		},
		{
			name: "pr_unlabeled by a collaborator",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionUnlabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
				Sender: github.User{Login: "collaborator"},
				Label:  github.Label{Name: DefaultNeedsTwsLabel},
			},
			changes: []github.PullRequestChange{
				{
					Filename: "README.md",
				},
			},
			IssueCommentsAdded: 1,
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			l := externalplugin.NewLogger().With("test", c.name)
			defer l.Sync()
			fc := fakegithub.NewFakeClient()
			a := fakeAliases{
				Aliases: repoowners.RepoAliases{
					"technical-writers": {
						"reviewer":  {},
						"reviewer2": {},
					}},
			}
			fc.Collaborators = []string{"collaborator"}
			p := PluginBackend{
				ghc: fc,
				oac: a,
				gcf: fakeClientFactory{},
			}
			fc.PullRequestChanges[c.event.PullRequest.Number] = c.changes
			fc.IssueLabelsExisting = c.IssueLabelsExisting
			fc.Reviews[c.event.PullRequest.Number] = c.Reviews
			err := p.handlePullRequest(l, c.event)
			if err != nil {
				t.Errorf("handlePullRequest() returned error: %v", err)
			}
			if got, want := len(fc.IssueLabelsAdded), c.IssueLabelsAdded; got != want {
				t.Errorf("case %s, IssueLabelsAdded mismatch - got %d, want %d.", c.name, got, want)
			}
			if got, want := len(fc.IssueLabelsRemoved), c.IssueLabelsRemoved; got != want {
				t.Errorf("case %s, IssueLabelsRemoved mismatch - got %d, want %d.", c.name, got, want)
			}
			if got, want := len(fc.IssueCommentsAdded), c.IssueCommentsAdded; got != want {
				t.Errorf("case %s, IssueCommentsAdded mismatch - got %d, want %d.", c.name, got, want)
			}
		})
	}
}

func Test_HandlePullRequestReview(t *testing.T) {
	testcases := []struct {
		name           string
		event          github.ReviewEvent
		assigneesAdded []string
		labelsExisting []string
		labelsAdded    int
		labelsRemoved  int
	}{
		{
			name: "not a submitted review",
			event: github.ReviewEvent{
				Action: github.ReviewActionDismissed,
			},
		},
		{
			name:           "pr review approved and assigned, remove label",
			assigneesAdded: []string{"org/repo#101:reviewer"},
			labelsExisting: []string{"org/repo#101:do-not-merge/missing-docs-review"},
			labelsRemoved:  1,
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateApproved,
					User:  github.User{Login: "Reviewer"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
		{
			name: "pr review made by author",
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateApproved,
					User:  github.User{Login: "pr-author"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
		{
			name: "pr approve not made by required reviewer",
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateApproved,
					User:  github.User{Login: "bad-reviewer"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
		{
			name:           "pr changes requested by a reviewer, assign a reviewer, add label",
			assigneesAdded: []string{"org/repo#101:reviewer"},
			labelsAdded:    1,
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateChangesRequested,
					User:  github.User{Login: "Reviewer"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
		{
			name:           "pr changes requested, label already present",
			assigneesAdded: []string{"org/repo#101:reviewer"},
			labelsExisting: []string{"org/repo#101:do-not-merge/missing-docs-review"},
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateChangesRequested,
					User:  github.User{Login: "Reviewer"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
					Labels: []github.Label{
						{
							Name: DefaultNeedsTwsLabel,
						},
					},
				},
			},
		},
		{
			name:           "pr approved, label already removed",
			assigneesAdded: []string{"org/repo#101:reviewer"},
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateApproved,
					User:  github.User{Login: "Reviewer"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			fc := fakegithub.NewFakeClient()
			l := externalplugin.NewLogger().With("test", c.name)
			defer l.Sync()
			a := fakeAliases{
				Aliases: repoowners.RepoAliases{
					"technical-writers": {
						"reviewer": {},
					},
				}}
			fc.Collaborators = []string{"reviewer"}
			fc.IssueLabelsExisting = c.labelsExisting
			p := PluginBackend{
				ghc: fc,
				oac: a,
				gcf: fakeClientFactory{},
			}
			err := p.handlePullRequestReview(l, c.event)
			if err != nil {
				t.Errorf("handlePullRequestReview() returned an error where it shouldn't: %v", err)
			}
			if got, want := len(fc.AssigneesAdded), len(c.assigneesAdded); got != want {
				t.Errorf("case %s, number of assignees is wrong. got %d, want %d", c.name, got, want)
			}
			if got, want := len(fc.IssueLabelsAdded), c.labelsAdded; got != want {
				t.Errorf("case %s, added a label where it shouldn't have been added. got %d want %d", c.name, got, want)
			}
			if got, want := len(fc.IssueLabelsRemoved), c.labelsRemoved; got != want {
				t.Errorf("case %s, didn't remove a label where it should have been removed. got %d want %d", c.name, got, want)
			}
		})
	}
}
