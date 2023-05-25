package main

import (
	"reflect"
	"testing"

	rt "github.com/kyma-project/test-infra/development/tools/pkg/rendertemplates"
)

func Test_findDuplicatedTargetFiles(t *testing.T) {
	tc := []struct {
		Name               string
		ExpectedDuplicates []string
		Templates          []*rt.TemplateConfig
	}{
		{
			Name:               "no duplicates, return nil",
			ExpectedDuplicates: nil,
			Templates: []*rt.TemplateConfig{
				{
					FromTo: []rt.FromTo{
						{
							To: "../../prow/jobs/test-infra/testfile.yaml",
						},
						{
							To: "../../prow/jobs/test-infra/testfile2.yaml",
						},
						{
							To: "../../prow/jobs/test-infra/testfile3.yaml",
						},
					},
				},
			},
		},
		{
			Name:               "single duplicated path, return list of duplicated files",
			ExpectedDuplicates: []string{"../../prow/jobs/test-infra/testfile.yaml"},
			Templates: []*rt.TemplateConfig{
				{
					FromTo: []rt.FromTo{
						{
							To: "../../prow/jobs/test-infra/testfile.yaml",
						},
						{
							To: "../../prow/jobs/test-infra/testfile.yaml",
						},
						{
							To: "../../prow/jobs/test-infra/testfile3.yaml",
						},
					},
				},
			},
		},
		{
			Name:               "multiple duplicates, return list of duplicated files",
			ExpectedDuplicates: []string{"../../prow/jobs/test-infra/testfile.yaml", "../../prow/jobs/test-infra/testfile3.yaml"},
			Templates: []*rt.TemplateConfig{
				{
					FromTo: []rt.FromTo{
						{
							To: "../../prow/jobs/test-infra/testfile.yaml",
						},
						{
							To: "../../prow/jobs/test-infra/testfile.yaml",
						},
						{
							To: "../../prow/jobs/test-infra/testfile3.yaml",
						},
						{
							To: "../../prow/jobs/test-infra/testfile3.yaml",
						},
					},
				},
			},
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			duplicates := findDuplicatedTargetFiles(c.Templates)

			if !reflect.DeepEqual(duplicates, c.ExpectedDuplicates) {
				t.Errorf("findDuplicatedTargetFiles(): Got %v, but expected %v", duplicates, c.ExpectedDuplicates)
			}
		})
	}
}
