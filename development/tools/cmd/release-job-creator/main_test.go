package main

import (
	"testing"
)

func TestCreateJobName(t *testing.T) {
	tests := []struct {
		typ      string
		rel      string
		folder   string
		expected string
	}{
		{
			"pre",
			"1.2",
			"tests-kubeless-tests",
			"pre-rel12-tests-kubeless-tests",
		},
		{
			"pre",
			"1.3",
			"tests-event-bus-tests",
			"pre-rel13-tests-event-bus-tests",
		},
	}

	for _, test := range tests {
		job := createJobName(test.typ, test.rel, test.folder)
		if job != test.expected {
			t.Errorf("expected: %s, got %s", test.expected, job)
		}
	}

}
