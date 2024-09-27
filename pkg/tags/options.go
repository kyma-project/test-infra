package tags

import (
	"fmt"
)

type TagOption func(o *Tagger) error

func DateFormat(format string) TagOption {
	return func(t *Tagger) error {
		t.Date = t.Time.Format(format)
		return nil
	}
}

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

func PRNumber(pr string) TagOption {
	return func(t *Tagger) error {
		if len(pr) == 0 {
			return fmt.Errorf("pr number cannot be empty")
		}
		t.PRNumber = pr
		return nil
	}
}
