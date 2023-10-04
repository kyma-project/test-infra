package tags

type TagOption func(o *Tagger)

func DateFormat(format string) TagOption {
	return func(t *Tagger) {
		t.Date = t.Time.Format(format)
	}
}

func CommitSHA(sha string) TagOption {
	return func(t *Tagger) {
		t.CommitSHA = sha
	}
}
