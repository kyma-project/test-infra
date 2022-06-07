package tags

import (
	"testing"
	"time"
)

func TestTagger_BuildTag(t *testing.T) {
	tc := []struct {
		template string
		expected string
		tag      Tag
	}{
		{
			template: `v{{ .Date }}-{{ .ShortSHA }}`,
			expected: "v20220602-abc1234",
			tag: Tag{
				ShortSHA: "abc1234",
				Time:     time.Now(),
				Date:     time.Date(2022, 06, 02, 01, 01, 01, 1, time.Local).Format("20060102"),
			},
		},
	}
	for _, c := range tc {
		tg := Tagger{tagTemplate: c.template}
		got, err := tg.BuildTag(c.tag)
		if err != nil {
			t.Errorf("error building tag: %v", err)
		}
		if got != c.expected {
			t.Errorf("Tags do not match\n. Expected %v, Got %v", c.expected, got)
		}
	}
}
