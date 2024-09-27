package tags

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"text/template"
	"time"
)

type Tagger struct {
	tags      []Tag
	CommitSHA string
	ShortSHA  string
	PRNumber  string
	Time      time.Time
	Date      string
}

func (tg *Tagger) Env(key string) string {
	return os.Getenv(key)
}

func (tg *Tagger) ParseTags() ([]Tag, error) {
	var parsed []Tag
	for _, tag := range tg.tags {
		if len(tag.Name) == 0 || len(tag.Value) == 0 {
			return nil, fmt.Errorf("tag name or value is empty, tag name: %s, tag value: %s", tag.Name, tag.Value)
		}
		tmpl, err := template.New("tag").Parse(tag.Value)
		if err != nil {
			return nil, err
		}
		buf := bytes.Buffer{}
		err = tmpl.Execute(&buf, tg)
		if err != nil {
			return nil, err
		}
		tag.Value = buf.String()
		err = tg.validateTag(tag)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, tag)
	}

	return parsed, nil
}

func (tg *Tagger) validateTag(tag Tag) error {
	if tag.Name == "default_tag" && len(tag.Validation) == 0 {
		return fmt.Errorf("default_tag validation is empty, tag: %s", tag.Value)
	}
	if tag.Validation != "" {
		// Verify PR default tag. Check if value starts with PR- and is followed by a number
		re := regexp.MustCompile(tag.Validation)
		match := re.FindAllString(tag.Value, -1)
		if match == nil {
			return fmt.Errorf("tag validation failed, tag: %s, validation: %s", tag.Value, tag.Validation)
		}
	}
	return nil
}

func NewTagger(tags []Tag, opts ...TagOption) (*Tagger, error) {
	now := time.Now()
	t := Tagger{
		tags: tags,
		Time: now,
		Date: now.Format("20060102"),
	}
	for _, o := range opts {
		err := o(&t)
		if err != nil {
			return nil, fmt.Errorf("error applying tag option: %w", err)
		}
	}
	return &t, nil
}
