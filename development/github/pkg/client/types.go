package client

import (
	"github.com/google/go-github/v40/github"
)

// IssueTransferredEvent represent GitHub.IssuesEvent for transferred action.
// It adds support for NewIssue and NewRepository json keys from GitHub webhook.
type IssueTransferredEvent struct {
	github.IssuesEvent
	Changes *TransferredChange `json:"changes,omitempty"`
}

func (i *IssueTransferredEvent) GetChanges() *TransferredChange {
	if i == nil {
		return nil
	}
	return i.Changes
}

func (t *TransferredChange) GetNewIssue() *github.Issue {
	if t == nil {
		return nil
	}
	return t.NewIssue
}

func (t *TransferredChange) GetNewRepository() *github.Repository {
	if t == nil {
		return nil
	}
	return t.NewRepository
}

// TransferredChange represents the changes when an issue, has been transferred.
type TransferredChange struct {
	NewIssue      *github.Issue      `json:"new_issue,omitempty"`
	NewRepository *github.Repository `json:"new_repository,omitempty"`
}
