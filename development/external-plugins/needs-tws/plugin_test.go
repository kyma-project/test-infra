package main

import (
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/github/fakegithub"
	"k8s.io/test-infra/prow/repoowners"
	"testing"
)

type fakeAliases struct {
	ownersAliases
	Aliases repoowners.RepoAliases
}

type fakeGitClientFactory struct {
	git.ClientFactory
}

type fakeRepoClient struct {
	git.RepoClient
}

func (f fakeAliases) LoadOwnersAliases(l *logrus.Entry, basedir, filename string) (repoowners.RepoAliases, error) {
	return f.Aliases, nil
}

func (f fakeRepoClient) Directory() string {
	return ""
}

func (f fakeGitClientFactory) ClientFor(org, repo string) (git.RepoClient, error) {
	return fakeRepoClient{}, nil
}

func Test_hasMarkdownChanges(t *testing.T) {
	fc := fakegithub.NewFakeClient()
	fc.Commits = map[string]github.RepositoryCommit{
		"hasChangesSha": {
			SHA: "hasChangesSha",
			Files: []github.CommitFile{
				{
					Filename: "README.md",
				},
			},
		},
		"noChangesSha": {
			SHA: "noChangesSha",
			Files: []github.CommitFile{
				{
					Filename: "something/else.go",
				},
			},
		},
		"changedSubMdSha": {
			SHA: "changedSubMdSha",
			Files: []github.CommitFile{
				{
					Filename: "something/child/markdown.md",
				},
			},
		},
		"changedMdNotForTwsSha": {
			SHA: "changedMdNotForTwsSha",
			Files: []github.CommitFile{
				{
					Filename: "not_for_tws.md",
				},
			},
		},
	}

	p := Plugin{
		ghc: fc,
	}
	testcases := []struct {
		SHA      string
		Expected bool
	}{
		{"hasChangesSha", true},
		{"noChangesSha", false},
		{"changedSubMdSha", true},
		//{"changedMdNotForTwsSha", false},
	}
	for _, c := range testcases {
		t.Run(c.SHA, func(t *testing.T) {
			result, err := p.hasMarkdownChanges("foo", "bar", c.SHA)
			if err != nil {
				t.Fatalf("hasMarkdownChanges() premature error: %v\n", err)
			}
			if result != c.Expected {
				t.Logf("Bad test result for testcase %s\n. Got %v, Expected %v", c.SHA, result, c.Expected)
				t.Fail()
			}
		})
	}
}

func Test_HandlePullRequest(t *testing.T) {
	SHA := "9448a2cb0a3915ac956685de8ffb3d4ef55fbc05"
	twsLabel := "org/repo#101:needs-tws-review"
	testcases := []struct {
		name                string
		event               github.PullRequestEvent
		commit              github.RepositoryCommit
		IssueLabelsAdded    []string
		IssueLabelsExisting []string
	}{
		{
			name:             "pr_opened add label",
			IssueLabelsAdded: []string{twsLabel},
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
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
			},
		},
		{
			name:                "pr_synchronize already has a label",
			IssueLabelsAdded:    []string{twsLabel},
			IssueLabelsExisting: []string{twsLabel},
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
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
			},
		},
		{
			name: "pr_synchronize do not add label",
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
			name: "pr_labelled",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionLabeled,
			},
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			l := logrus.WithField("test", c.name)
			fc := fakegithub.NewFakeClient()
			p := Plugin{
				ghc: fc,
			}
			fc.Commits[SHA] = c.commit
			fc.IssueLabelsAdded = c.IssueLabelsExisting

			err := p.handlePullRequest(l, c.event)
			if err != nil {
				t.Errorf("handlePullRequest() returned error: %v", err)
			}
			if got, want := len(fc.IssueLabelsAdded), len(c.IssueLabelsAdded); got != want {
				t.Errorf("case %s, IssueLabelsAdded mismatch - got %d, want %d.", c.name, got, want)
			}
		})
	}
}

func Test_HandlePullRequestReview(t *testing.T) {
	testcases := []struct {
		name      string
		event     github.ReviewEvent
		assignees []string
		labels    []string
	}{
		{
			name: "not a submitted review",
			event: github.ReviewEvent{
				Action: github.ReviewActionDismissed,
			},
		},
		{
			name:      "pr review approved and assigned, remove label",
			assignees: []string{"org/repo#101:reviewer"},
			labels:    []string{"org/repo#101:needs-tws-review"},
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateApproved,
					User:  github.User{Login: "reviewer"},
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
			name:      "pr changes requested by a reviewer, assign a reviewer",
			assignees: []string{"org/repo#101:reviewer"},
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateChangesRequested,
					User:  github.User{Login: "reviewer"},
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
			l := logrus.WithField("test", c.name)
			a := fakeAliases{
				Aliases: repoowners.RepoAliases{
					"technical-writers": {
						"reviewer": {},
					},
				}}
			fc.Collaborators = []string{"reviewer"}
			fc.IssueLabelsExisting = c.labels
			p := Plugin{
				ghc: fc,
				oac: a,
				gcf: fakeGitClientFactory{},
			}
			err := p.handlePullRequestReview(l, c.event)
			if err != nil {
				t.Errorf("handlePullRequestReview() returned an error where it shouldn't: %v", err)
			}
			if got, want := len(fc.AssigneesAdded), len(c.assignees); got != want {
				t.Errorf("case %s, number of assignees is wrong. got %d, want %d", c.name, got, want)
			}
			if got, want := len(fc.IssueLabelsRemoved), len(c.labels); got != want {
				t.Errorf("case %s, didn't remove a label where it should have been removed. got %d want %d", c.name, got, want)
			}
		})
	}
}
