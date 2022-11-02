package tags

import (
	"testing"
	"time"
)

func TestOption_CommitSHA(t *testing.T) {
	tc := struct {
		sha      string
		expected string
	}{
		sha:      "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
	}
	tag := Tagger{
		Time:      time.Now(),
		CommitSHA: "1edd8d99e07c726c2226713312ae9551162b825b",
	}
	f := CommitSHA(tc.sha)
	f(&tag)
	if tag.CommitSHA != tc.expected {
		t.Errorf("%s != %s", tag.CommitSHA, tc.expected)
	}
}

func TestOption_DateFormat(t *testing.T) {
	now := time.Now()
	tc := struct {
		dateFormat   string
		expectedDate string
	}{
		dateFormat:   "2006-01-02",
		expectedDate: now.Format("2006-01-02"),
	}
	tag := Tagger{
		Time: now,
	}
	f := DateFormat(tc.dateFormat)
	f(&tag)
	if tag.Date != tc.expectedDate {
		t.Errorf("%s != %s", tag.Date, tc.expectedDate)
	}
}
