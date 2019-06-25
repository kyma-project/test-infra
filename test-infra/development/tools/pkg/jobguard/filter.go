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
	}, nil
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

// IsFailedStatus returns true for status that have Error of Failure state
func IsFailedStatus(status string) bool {
	return status == StatusStateError || status == StatusStateFailure
}

// IsPendingStatus returns true for status that are in Pending state
func IsPendingStatus(status string) bool {
	return status == StatusStatePending
}
