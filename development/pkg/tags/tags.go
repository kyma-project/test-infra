package tags

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"text/template"
	"time"
)

type Tagger struct {
	tags                []Tag
	CommitSHA, ShortSHA string
	Time                time.Time
	Date                string
}

// TODO (@Ressetkk): Evaluate if we really need to implement it in central way
//func (tg *Tagger) AddFlags(fs *flag.FlagSet) {
//	fs.Var(&tg.tags, "tag", "Go-template based tag")
//}

func (tg *Tagger) Env(key string) string {
	return os.Getenv(key)
}

func (tg *Tagger) ParseTags() ([]Tag, error) {
	var parsed []Tag
	for _, t := range tg.tags {
		tmpl, err := template.New("tag").Parse(t.Value)
		if err != nil {
			return nil, err
		}
		buf := bytes.Buffer{}
		err = tmpl.Execute(&buf, tg)
		if err != nil {
			return nil, err
		}
		tag := Tag{
			Name:  t.Name,
			Value: buf.String(),
		}
		parsed = append(parsed, tag)
	}

	return parsed, nil
}

func NewTagger(tags []Tag, opts ...TagOption) (*Tagger, error) {
	now := time.Now()
	t := Tagger{
		tags: tags,
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
