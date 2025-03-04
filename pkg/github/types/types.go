package types

import (
	"github.com/google/go-github/v48/github"
	"github.com/kyma-project/test-infra/pkg/types"
)

// SearchIssuesResult hnews information if issue was found and list found GitHub Issues.
// Fields names are meaningfully so are easy to use in composition types.
type SearchIssuesResult struct {
	GithubIssueFound *bool           `json:"githubIssueFound"`
	GithubIssues     []*github.Issue `json:"githubIssuesReport,omitempty"`
}

// IssueMetadata hnews metadata about GitHub Issue.
// Fields names are meaningfully so are easy to use in composition types.
type IssueMetadata struct {
	GithubIssueOrg      *string    `json:"githubIssueOrg,omitempty"`
	GithubIssueRepo     *string    `json:"githubIssueRepo,omitempty"`
	GithubIssueNumber   *int       `json:"githubIssueNumber,omitempty"`
	GithubIssueURL      *string    `json:"githubIssueURL,omitempty"`
	GithubIssueAssignee types.User `json:"githubIssueAssignee,omitempty"`
}
# (2025-03-04)