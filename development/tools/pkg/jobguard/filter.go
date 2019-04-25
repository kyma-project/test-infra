package jobguard

import (
	"regexp"
)

// StatusPredicate defines signature of a function that checks if a Status fulfil some criteria
type StatusPredicate func(in Status) bool

// NameRegexpPredicate returns function that checks if status name matches pattern
func NameRegexpPredicate(pattern string) (StatusPredicate, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(in Status) bool {
		return r.MatchString(in.Name)
	},nil
}

// FailedStatusPredicate returns true for Statuses that have Error of Failure state
func FailedStatusPredicate(in Status) bool {
	return in.State == string(StatusStateError) || in.State == string(StatusStateFailure)
}

// PendingStatusPredicate returns true for Statuses that are in Pending state
func PendingStatusPredicate(in Status) bool {
	return in.State == string(StatusStatePending)
}

// Filter removes Statused that do not match predicate
func Filter(in []Status, pred StatusPredicate) []Status {
	var filteredStatuses []Status

	for _, s := range in {
		if pred(s) {
			filteredStatuses = append(filteredStatuses, s)
		}
	}
	return filteredStatuses
}