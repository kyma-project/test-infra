package tags

type TagOption func(o *Tag)

func DateFormat(format string) TagOption {
	return func(t *Tag) {
		t.Date = t.Time.Format(format)
	}
}

func CommitSHA(sha string) TagOption {
	return func(t *Tag) {
		t.CommitSHA = sha
	}
}
