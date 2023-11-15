package main

import (
	"fmt"
	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeCLAServer struct {
	org, repo string
	number    int
	requested int
}

func (f *fakeCLAServer) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// This is roughly based on a response we get from https://cla-assistant.io/check endpoint
	// It's funny enough because CLA Assistant does not make any checks before running this endpoint.
	// It just redirects to https://github.com, lol.
	w.Header().Set("Location", "https://github.com")
	w.WriteHeader(302)
	fmt.Fprintf(w, "Found. Redirecting to https://github.com")
	f.requested++
}

func Test_handleGenericCommentEvent(t *testing.T) {
	testcases := []struct {
		name             string
		expectedRequests int
		event            github.IssueCommentEvent
	}{
		{
			name:             "Non-PR comment created, skip CLA recheck",
			expectedRequests: 0,
			event: github.IssueCommentEvent{
				Action: github.IssueCommentActionCreated,
				Issue:  github.Issue{PullRequest: nil},
			},
		},
		{
			name:             "Non-created event received, skip CLA recheck",
			expectedRequests: 0,
			event: github.IssueCommentEvent{
				Action: github.IssueCommentActionEdited,
				Issue:  github.Issue{PullRequest: &struct{}{}},
			},
		},
		{
			name:             "PR comment created, without /cla-recheck, skip CLA recheck",
			expectedRequests: 0,
			event: github.IssueCommentEvent{
				Action:  github.IssueCommentActionCreated,
				Issue:   github.Issue{PullRequest: &struct{}{}},
				Comment: github.IssueComment{Body: ""},
			},
		},
		{
			name:             "PR comment created, /cla-recheck, request CLA check",
			expectedRequests: 1,
			event: github.IssueCommentEvent{
				Action:  github.IssueCommentActionCreated,
				Issue:   github.Issue{PullRequest: &struct{}{}, Number: 101},
				Comment: github.IssueComment{Body: "/cla-recheck"},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fclas := fakeCLAServer{}
			s := httptest.NewServer(&fclas)
			defer s.Close()
			clac := CLAClient{
				address: s.URL,
				client:  http.Client{Timeout: time.Minute * 5},
			}
			l := externalplugin.NewLogger().With("test", tc.name)
			if err := clac.handleIssueCommentEvent(l, tc.event); err != nil {
				t.Errorf("Unexpected error from handleIssueCommentEvent: %v", err)
			}
			if fclas.requested != tc.expectedRequests {
				t.Errorf("Wrong number of requested CLA rechecks. Got %d, Wanted %d", fclas.requested, tc.expectedRequests)
			}
		})
	}
}

func Test_helpProvider(t *testing.T) {
	th, err := helpProvider([]config.OrgRepo{})
	if err != nil {
		t.Errorf("unexpected error from helpProvider %v", err)
	}
	expCmds := 1
	if len(th.Commands) != expCmds {
		t.Errorf("Number of commands does not match Got %d Wanted %d", len(th.Commands), expCmds)
	}
}
