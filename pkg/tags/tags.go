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

// TODO (@Ressetkk): Evaluate if we really need to implement it in central way
// func (tg *Tagger) AddFlags(fs *flag.FlagSet) {
//	fs.Var(&tg.tags, "tag", "Go-template based tag")
// }

func (tg *Tagger) Env(key string) string {
	return os.Getenv(key)
}

func (tg *Tagger) ParseTags() ([]Tag, error) {
	var parsed []Tag
	for _, t := range tg.tags {
		if len(t.Name) == 0 || len(t.Value) == 0 {
			return nil, fmt.Errorf("tag name or value is empty, tag name: %s, tag value: %s", t.Name, t.Value)
		}
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
		err = tg.validateTag(tag)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, tag)
	}

	return parsed, nil
}

func (tg *Tagger) validateTag(tag Tag) error {
	if tag.Name == "default_tag" && tag.Validation == "" {
		return fmt.Errorf("default_tag validation is empty")
	}
	if tag.Validation != "" {
		// Verify PR default tag. Check if value starts with PR- and is followed by a number
		re := regexp.MustCompile(tag.Validation)
		match := re.FindAllString(tag.Value, -1)
		if match == nil {
			return fmt.Errorf("default_tag validation failed")
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
		o(&t)
	}
	// TODO (dekiel): this should be valideted outside of constructor.
	//  The tagger should be able to work for pr default tag or commit default tag.
	//  The different default tags require different values.
	// if t.CommitSHA == "" {
	// 	return nil, errors.New("variable CommitSHA is empty")
	// }
	// t.ShortSHA = fmt.Sprintf("%.8s", t.CommitSHA)
	return &t, nil
}
