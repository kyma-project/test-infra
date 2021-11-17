package client

import (
	"github.com/google/go-github/v40/github"
)

type IssueTransferred struct {
	github.IssuesEvent
	// Action is the action that was performed. Possible values are: "opened",
	// "edited", "deleted", "transferred", "pinned", "unpinned", "closed", "reopened",
	// "assigned", "unassigned", "labeled", "unlabeled", "locked", "unlocked",
	// "milestoned", or "demilestoned".
	//Action   *string `json:"action,omitempty"`
	//Issue    *github.Issue  `json:"issue,omitempty"`
	//Assignee *github.User   `json:"assignee,omitempty"`
	//Label    *github.Label  `json:"label,omitempty"`

	// The following fields are only populated by Webhook events.
	Changes *TransferredChange `json:"changes,omitempty"`
	//Repo         *github.Repository   `json:"repository,omitempty"`
	//Sender       *github.User         `json:"sender,omitempty"`
	//Installation *github.Installation `json:"installation,omitempty"`
}

func (i *IssueTransferred) GetChanges() *TransferredChange {
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

// EditChange represents the changes when an issue, pull request, or comment has
// been edited.
type TransferredChange struct {
	NewIssue      *github.Issue      `json:"new_issue,omitempty"`
	NewRepository *github.Repository `json:"new_repository,omitempty"`
}

//if *event.Action == "transferred" {
//transferredPayload := &IssueTransferred{}
//p := (*json.RawMessage)(&payload)
//err := json.Unmarshal(*p, &transferredPayload)
//if err != nil { fmt.Printf("failed unmarshal json, %s\n", err)}
//fmt.Printf("%v\n", transferredPayload)
//}

//t := fmt.Sprintf("sap.kyma.custom.%s.%s.v1", k.appName, eventType)
//kymaEventingType := strings.Replace(t, "-", "", -1)

//event.SetType(kymaEventingType)
