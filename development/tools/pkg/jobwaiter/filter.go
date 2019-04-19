package jobwaiter

import "strings"

// FilterStatusByName filters statuses by name
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

// FailedStatuses filters statuses by failed or error state
func FailedStatuses(in []Status) []Status {
	var filteredStatuses []Status

	for _, s := range in {
		if s.State != string(StatusStateError) && s.State != string(StatusStateFailure) {
			continue
		}

		filteredStatuses = append(filteredStatuses, s)
	}

	return filteredStatuses
}

// PendingStatuses filters statuses by pending state
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
