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
	FailOnNoContexts bool
	PredicateFunc    StatusPredicate

	Org     string
	Repo    string
	BaseRef string
}

type StatusMap map[string]string
type StatusPredicate func(in github.Status) bool

func RegexpPredicate(p string) StatusPredicate {
	expr := regexp.MustCompile(p)
	return func(in github.Status) bool {
		return expr.MatchString(in.Context)
	}
}

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

func (sm StatusMap) PendingList() string {
	var list string
	for k, v := range sm {
		if v == StatusPending {
			list += k + " "
		}
	}
	return list
}

func (sm StatusMap) FailedList() string {
	var list string
	for k, v := range sm {
		if v == StatusFailure || v == StatusError {
			list += k + " "
		}
	}
	return list
}

func (sm StatusMap) String() string {
	var res string
	for s := range sm {
		res += s + " "
	}
	return res
}
