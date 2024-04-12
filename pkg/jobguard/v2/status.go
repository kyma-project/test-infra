package v2

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"k8s.io/test-infra/prow/github"
)

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusPending = "pending"
	StatusError   = "error"
)

type StatusOptions struct {
	FailOnNoContexts bool
	PredicateFunc    StatusPredicate

	Org     string
	Repo    string
	BaseRef string
}

type StatusMap map[string]string
type StatusPredicate func(in github.Status) bool

// RegexpPredicate returns function that checks if status name matches the pattern.
func RegexpPredicate(p string) StatusPredicate {
	expr := regexp.MustCompile(p)
	return func(in github.Status) bool {
		return expr.MatchString(in.Context)
	}
}

// FetchRequiredStatuses fetches statuses of GitHub commit,
// then builds the StatusMap by a given StatusPredicate with their initial state.
func (so StatusOptions) FetchRequiredStatuses(c github.Client, sp StatusPredicate) (StatusMap, error) {
	r, err := c.GetCombinedStatus(so.Org, so.Repo, so.BaseRef)
	if err != nil {
		return nil, fmt.Errorf("status fetch error: %w", err)
	}

	required := make(StatusMap)
	for _, st := range r.Statuses {
		if sp(st) {
			required[st.Context] = st.State
		}
	}
	if so.FailOnNoContexts && len(required) == 0 {
		return nil, errors.New("no statuses math the given predicate")
	}
	return required, nil
}

// Update updates the source StatusMap with the new states of the statuses.
func (so StatusOptions) Update(c github.Client, required StatusMap) (StatusMap, error) {
	r, err := c.GetCombinedStatus(so.Org, so.Repo, so.BaseRef)
	if err != nil {
		return nil, fmt.Errorf("status fetch error: %w", err)
	}

	for _, st := range r.Statuses {
		if _, ok := required[st.Context]; ok {
			required[st.Context] = st.State
		}
	}

	return required, nil
}

// CombinedStatus returns the overall state of the contexts.
// If one of the statuses is in "pending" state then it assumes tht the overall status is still "pending".
// If one of the statuses is in "failed" state then the entire context is "failed".
// If none of the above is valid then context has "success" state.
func (sm StatusMap) CombinedStatus() string {
	status := StatusSuccess
	for _, v := range sm {
		if v == StatusPending {
			status = v
			break
		}
		if v == StatusFailure || v == StatusError {
			status = StatusFailure
			break
		}
	}
	return status
}

// PendingList returns list of statuses with "pending" state.
func (sm StatusMap) PendingList() string {
	var list string
	for k, v := range sm {
		if v == StatusPending {
			list += k + " "
		}
	}
	return list
}

// FailedList returns list of statuses with "failure" or "error" states.
func (sm StatusMap) FailedList() string {
	var list string
	for k, v := range sm {
		if v == StatusFailure || v == StatusError {
			list += k + " "
		}
	}
	return list
}

// String implements Stringer interface
func (sm StatusMap) String() string {
	var res string
	for s := range sm {
		res += s + " "
	}
	return res
}
