package tags

import (
	"bytes"
	"flag"
	"text/template"
	"time"
)

type Tagger struct {
	tagTemplate string
}

type Tag struct {
	CommitSHA, ShortSHA string
	Time                time.Time
	Date                string
}

func (t *Tagger) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&t.tagTemplate, "tag-template", `v{{ .Date }}-{{ .ShortSHA }}`, "Go-template based tag")
}

func (t Tagger) BuildTag(s Tag) (string, error) {
	tmpl, err := template.New("tag").Parse(t.tagTemplate)
	if err != nil {
		return "", err
	}
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, s)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

type TagOption func(o *Tag)

func DateFormat(format string) TagOption {
	return func(t *Tag) {
		t.Date = t.Time.Format(format)
	}
}

func NewTag(opts ...TagOption) {
	now := time.Now()
	t := Tag{
		Time: now,
		Date: now.Format("20060102"),
	}

	for _, o := range opts {
		o(&t)
	}
}
