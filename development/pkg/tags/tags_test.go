package tags

import (
	"testing"
	"time"
)

func TestTagger_ParseTags(t *testing.T) {
	tc := []struct {
		name     string
		template []Tag
		expected Tag
		expErr   bool
	}{
		{
			name:     "tag is v20220602-abc1234",
			template: []Tag{{Name: "TagTemplate", Value: `v{{ .Date }}-{{ .ShortSHA }}`}},
			expected: Tag{Name: "TagTemplate", Value: "v20220602-abc1234"},
		},
		{
			name:     "fail, malformed tag template",
			template: []Tag{{Name: "", Value: "{{ ..Date }}-{{ .ShortSHA }}"}},
			expErr:   true,
			expected: Tag{},
		},
		{
			name:     "could not parse, missing field",
			template: []Tag{{Name: "", Value: "{{ .Missing }}-field"}},
			expected: Tag{},
			expErr:   true,
		},
		{
			name:     "tag is v20220602-test",
			template: []Tag{{Name: "Test", Value: `v{{ .Date }}-{{ .Env "test-var" }}`}},
			expected: Tag{Name: "Test", Value: "v20220602-test"},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("test-var", "test")
			tag := Tagger{
				tags:      c.template,
				ShortSHA:  "abc1234",
				CommitSHA: "f1c7ca0b532141898f56c1843ae60ebec3a75a85",
				Time:      time.Now(),
				Date:      time.Date(2022, 06, 02, 01, 01, 01, 1, time.Local).Format("20060102"),
			}
			got, err := tag.ParseTags()
			if err != nil {
				if !c.expErr {
					t.Errorf("BuildTag caught error but didn't want to: %v", err)
				}
			} else {
				for _, tt := range got {
					if tt != c.expected {
						t.Errorf("expected %v != got %v", c.expected, got)
					}
				}
			}
		})
	}
}
