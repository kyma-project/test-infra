package tags

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/kyma-project/test-infra/pkg/logging"
	"go.uber.org/zap"
)

type Logger interface {
	logging.StructuredLoggerInterface
	logging.WithLoggerInterface
}

type Tagger struct {
	tags      []Tag
	logger    Logger
	CommitSHA string
	ShortSHA  string
	PRNumber  string
	Time      time.Time
	Date      string
}

func (tg *Tagger) Env(key string) string {
	tg.logger.Debugw("reading environment variable", "variable_name", key, "variable_value", os.Getenv(key))
	return os.Getenv(key)
}

func (tg *Tagger) ParseTags() ([]Tag, error) {
	tg.logger.Debugw("started parsing tags")
	var parsed []Tag
	for _, tag := range tg.tags {
		if len(tag.Name) == 0 || len(tag.Value) == 0 {
			return nil, fmt.Errorf("tag name or value is empty, tag name: %s, tag value: %s", tag.Name, tag.Value)
		}
		logger := tg.logger.With("tag", tag.Name, "value", tag.Value)
		logger.Debugw("verified tag name and value are not empty")
		logger.Debugw("parsing tag template")
		tmpl, err := template.New("tag").Parse(tag.Value)
		if err != nil {
			return nil, fmt.Errorf("error parsing tag template: %w", err)
		}
		logger.Debugw("parsed tag template")
		buf := bytes.Buffer{}
		err = tmpl.Execute(&buf, tg)
		if err != nil {
			return nil, fmt.Errorf("error executing tag template: %w", err)
		}
		logger.Debugw("successfully executed tag template", "computed_name", tag.Name, "computed_value", buf.String())
		tag.Value = buf.String()
		err = tg.validateTag(tag)
		if err != nil {
			return nil, fmt.Errorf("failed to validate tag: %w", err)
		}
		logger.Debugw("tag validation passed")
		parsed = append(parsed, tag)
		logger.Debugw("added tag to parsed tags")
	}
	tg.logger.Debugw("all tags parsed", "parsed_tags", parsed)

	return parsed, nil
}

func (tg *Tagger) validateTag(tag Tag) error {
	logger := tg.logger.With("tag", tag.Name, "value", tag.Value, "validation", tag.Validation)
	logger.Debugw("started validating tag")
	logger.Debugw("checking if validation regex is provided")
	if tag.Name == "default_tag" && len(tag.Validation) == 0 {
		return fmt.Errorf("default_tag required validation is empty, tag: %s", tag.Value)
	}
	if tag.Validation != "" {
		re := regexp.MustCompile(tag.Validation)
		logger.Debugw("compiled regex", "regex", re.String())
		match := re.FindAllString(tag.Value, -1)
		if match == nil {
			return fmt.Errorf("no regex match found, tag: %s, validation: %s", tag.Value, tag.Validation)
		}
		logger.Debugw("regex matched successfully", "match", match)
	}
	return nil
}

func NewTagger(logger Logger, tags []Tag, opts ...TagOption) (*Tagger, error) {
	logger.Debugw("started creating new tagger", "tags", tags)
	now := time.Now()
	logger.Debugw("read current time", "time", now)
	t := Tagger{
		tags: tags,
		Time: now,
		Date: now.Format("20060102"),
	}
	logger.Debugw("created new tagger, applying options")
	for _, o := range opts {
		err := o(&t)
		if err != nil {
			return nil, fmt.Errorf("error applying tag option: %w", err)
		}
	}
	logger.Debugw("applied options to tagger")
	if t.logger == nil {
		t.logger = zap.NewNop().Sugar()
		logger.Debugw("no logger provided for tagger, used nop logger")
	}
	logger.Debugw("finished creating new tagger")
	return &t, nil
}
# (2025-03-04)