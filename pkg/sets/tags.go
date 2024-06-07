package sets

import (
	"fmt"
	"strings"

	"github.com/kyma-project/test-infra/pkg/tags"
)

type Tags []tags.Tag

func (t *Tags) Set(val string) error {
	tg, err := tags.NewTagFromString(val)
	if err != nil {
		return err
	}
	*t = append(*t, tg)

	return nil
}

func (t *Tags) String() string {
	var stringTags []string

	for _, tg := range *t {
		stringTags = append(stringTags, fmt.Sprintf("%s=%s", tg.Name, tg.Value))
	}

	return strings.Join(stringTags, ",")
}

func (t *Tags) StringOnlyValues() string {
	var stringTags []string

	for _, tg := range *t {
		stringTags = append(stringTags, tg.Value)
	}

	return strings.Join(stringTags, ",")
}
