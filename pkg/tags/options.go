package tags

import (
	"fmt"
)

type TagOption func(o *Tagger)

func DateFormat(format string) TagOption {
	return func(t *Tagger) {
		t.Date = t.Time.Format(format)
	}
}

func CommitSHA(sha string) TagOption {
	return func(t *Tagger) {
		t.CommitSHA = sha
		// TODO (dekiel): This should be logged as a warning, not considered as an error
		// if t.CommitSHA == "" {
		// 	return nil, errors.New("variable CommitSHA is empty")
		// }
		t.ShortSHA = fmt.Sprintf("%.8s", t.CommitSHA)
	}
}

func PRNumber(pr string) TagOption {
	return func(t *Tagger) {
		t.PRNumber = pr
		// TODO (dekiel): The empty string should be logged as a warning.
	}
}
