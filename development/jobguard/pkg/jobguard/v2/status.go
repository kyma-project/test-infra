package v2

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/test-infra/prow/github"
	"regexp"
)

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusPending = "pending"
	StatusError   = "error"
)

type StatusOptions struct {
	ExpContextRegexp string
	FailOnNoContexts bool

	Org     string
	Repo    string
	BaseRef string
}

type Status struct {
	Context string
	State   string
}

func (so StatusOptions) BuildStatuses(c github.Client) ([]Status, error) {
	pattern, err := regexp.Compile(so.ExpContextRegexp)
	if err != nil {
		return nil, fmt.Errorf("regexp error: %w", err)
	}
	statuses, err := c.GetCombinedStatus(so.Org, so.Repo, so.BaseRef)
	if err != nil {
		return nil, fmt.Errorf("status fetch error: %w", err)
	}

	var expectedStatuses []Status
	for _, s := range statuses.Statuses {
		if pattern.MatchString(s.Context) {
			expectedStatuses = append(expectedStatuses, Status{s.Context, s.State})
		}
	}
	if so.FailOnNoContexts {
		return nil, errors.New("context is empty")
	}
	return expectedStatuses, nil
}

func (so StatusOptions) Update(c github.Client, requiredStatuses []Status) ([]Status, error) {
	statuses, err := c.GetCombinedStatus(so.Org, so.Repo, so.BaseRef)
	if err != nil {
		return nil, fmt.Errorf("status fetch error: %w", err)
	}

	for _, s := range statuses.Statuses {
		for _, cs := range requiredStatuses {
			if cs.Equals(s.Context) {
				cs.State = s.State
				break // we've updated the status, no need to check further
			}
		}
	}

	return requiredStatuses, nil
}

func (st Status) Equals(s string) bool {
	return st.Context == s
}

func (st Status) IsFailed() bool {
	return st.State == StatusError || st.State == StatusFailure
}

func (st Status) IsSuccess() bool {
	return st.State == StatusSuccess
}

func (st Status) IsPending() bool {
	return st.State == StatusPending
}
