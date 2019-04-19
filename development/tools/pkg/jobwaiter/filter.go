package jobwaiter

import "strings"

func FilterStatusByName(in []Status, substring string) []Status {
	var filteredStatuses []Status

	for _, s := range in {
		if !strings.Contains(s.Name, substring) {
			continue
		}

		filteredStatuses = append(filteredStatuses, s)
	}

	return filteredStatuses
}

func FailedStatuses(in []Status) []Status {
	var filteredStatuses []Status

	for _, s := range in {
		if s.State != string(StatusStateError) || s.State != string(StatusStateFailure) {
			continue
		}

		filteredStatuses = append(filteredStatuses, s)
	}

	return filteredStatuses
}

func PendingStatuses(in []Status) []Status {
	var filteredStatuses []Status

	for _, s := range in {
		if s.State != string(StatusStatePending) {
			continue
		}

		filteredStatuses = append(filteredStatuses, s)
	}

	return filteredStatuses
}
