package tags

import (
	"fmt"
)

type TagOption func(o *Tagger) error

// DateFormat sets Tagger Date field to the given value.
// It returns an error if the given value is empty.
func DateFormat(format string) TagOption {
	return func(t *Tagger) error {
		if len(format) == 0 {
			return fmt.Errorf("date format cannot be empty")
		}
		t.Date = t.Time.Format(format)
		return nil
	}
}

// CommitSHA sets Tagger CommitSHA field to the given value.
// It also sets the Tagger ShortSHA field to the first 8 characters of the given value.
// It returns an error if the given value is empty.
func CommitSHA(sha string) TagOption {
	return func(t *Tagger) error {
		if len(sha) == 0 {
			return fmt.Errorf("sha cannot be empty")
		}
		t.CommitSHA = sha
		t.ShortSHA = fmt.Sprintf("%.8s", t.CommitSHA)
		return nil
	}
}

// PRNumber sets Tagger PRNumber field to given value.
// It returns error if given value is empty.
func PRNumber(pr string) TagOption {
	return func(t *Tagger) error {
		if len(pr) == 0 {
			return fmt.Errorf("pr number cannot be empty")
		}
		t.PRNumber = pr
		return nil
	}
}
