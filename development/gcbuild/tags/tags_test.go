package tags

import (
	"flag"
	"reflect"
	"testing"
	"time"
)

func TestTagger_BuildTag(t *testing.T) {
	tc := []struct {
		name     string
		template string
		expected string
		expErr   bool
	}{
		{
			name:     "tag is v20220602-abc1234",
			template: `v{{ .Date }}-{{ .ShortSHA }}`,
			expected: "v20220602-abc1234",
		},
		{
			name:     "fail, malformed tag template",
			template: "{{ ..Date }}-{{ .ShortSHA }}",
			expErr:   true,
			expected: "",
		},
		{
			name:     "could not parse, missing field",
			template: "{{ .Missing }}-field",
			expected: "",
			expErr:   true,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			tag := Tag{
				ShortSHA:  "abc1234",
				CommitSHA: "f1c7ca0b532141898f56c1843ae60ebec3a75a85",
				Time:      time.Now(),
				Date:      time.Date(2022, 06, 02, 01, 01, 01, 1, time.Local).Format("20060102"),
			}
			tg := Tagger{TagTemplate: c.template}
			got, err := tg.BuildTag(&tag)
			if err != nil {
				if !c.expErr {
					t.Errorf("BuildTag caught error but didn't want to: %v", err)
				}
			} else {
				if got != c.expected {
					t.Errorf("expected %v != got %v", c.expected, got)
				}
			}
		})
	}
}

func TestTagger_AddFlags(t *testing.T) {
	tc := []struct {
		name         string
		flags        []string
		expectTagger Tagger
	}{
		{
			name:         "default tag template",
			expectTagger: Tagger{TagTemplate: `v{{ .Date }}-{{ .ShortSHA }}`},
		},
		{
			name:         "custom tag template",
			expectTagger: Tagger{TagTemplate: `{{ .CommitSHA }}`},
			flags:        []string{`--tag-template={{ .CommitSHA }}`},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			fs := flag.NewFlagSet("tag", flag.ContinueOnError)
			tr := Tagger{}
			tr.AddFlags(fs)
			err := fs.Parse(c.flags)
			if err != nil {
				t.Errorf("BuildTag caught error but didn't want to: %v", err)
			}
			if !reflect.DeepEqual(tr, c.expectTagger) {
				t.Errorf("expected %v != got %v", tr, c.expectTagger)
			}
		})
	}

}
