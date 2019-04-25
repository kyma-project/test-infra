package jobguard

import (
	"regexp"
)

type StatusPredicate func(in Status) bool

func NameRegexpPredicate(pattern string) (StatusPredicate, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(in Status) bool {
		return r.MatchString(in.Name)
	},nil
}

func FailedStatusPredicate(in Status) bool {
	return in.State == string(StatusStateError) || in.State == string(StatusStateFailure)
}

func PendingStatusPredicate(in Status) bool {
	return in.State == string(StatusStatePending)
}
func Filter(in []Status, pred StatusPredicate) []Status {
	var filteredStatuses []Status

	for _, s := range in {
		if pred(s) {
			filteredStatuses = append(filteredStatuses, s)
		}
	}
	return filteredStatuses
}