package client

import "github.com/google/go-github/v42/github"

type SearchIssuesResult struct {
	IssueFound *bool           `json:"issueFound"`
	Issues     []*github.Issue `json:"issuesReport,omitempty"`
}
