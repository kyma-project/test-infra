package tags

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"text/template"
	"time"
)

type Tagger struct {
	TagTemplate string
}

type Tag struct {
	CommitSHA, ShortSHA string
	Time                time.Time
	Date                string
}

func (t *Tagger) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&t.TagTemplate, "tag-template", `v{{ .Date }}-{{ .ShortSHA }}`, "Go-template based tag")
}

func (t Tagger) BuildTag(s *Tag) (string, error) {
	tmpl, err := template.New("tag").Parse(t.TagTemplate)
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

func (t Tag) Env(key string) string {
	return os.Getenv(key)
}

func NewTag(opts ...TagOption) (*Tag, error) {
	now := time.Now()
	t := Tag{
		Time: now,
		Date: now.Format("20060102"),
	}
	for _, o := range opts {
		o(&t)
	}
	if t.CommitSHA == "" {
		return nil, errors.New("variable CommitSHA is empty")
	}
	t.ShortSHA = fmt.Sprintf("%.8s", t.CommitSHA)
	return &t, nil
}
