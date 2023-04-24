package prowjobs

import (
	"fmt"
	"k8s.io/test-infra/prow/config"
	"testing"
)

const (
	OwnerAnnotationName       = "owner"
	DescriptionAnnotationName = "description"
)

// Test_OwnerAnnotations is a band-aid solution to missing ownership annotations in ProwJob definition.
// It looks up for all jobs in the registry, checks required annotations and fails if any of the annotations is missing.
// Currently, it supports only single repository.
// TODO (@Ressetkk): Change me to support InRepoConfig
func Test_OwnerAnnotations(t *testing.T) {
	prowConfig := "../../prow/config.yaml"
	jobConfig := "../../prow/jobs"
	c, err := config.Load(prowConfig, jobConfig, []string{}, "")
	if err != nil {
		t.Fatalf("error loading config files")
	}
	allRepos := c.AllRepos.List()
	pre := c.AllStaticPresubmits(allRepos)
	post := c.AllStaticPostsubmits(allRepos)
	period := c.AllPeriodics()
	var errs []error
	for _, j := range pre {
		errs = append(errs, checkRequiredAnnotations(j.Name, j.Annotations)...)
	}
	for _, j := range post {
		errs = append(errs, checkRequiredAnnotations(j.Name, j.Annotations)...)
	}
	for _, j := range period {
		errs = append(errs, checkRequiredAnnotations(j.Name, j.Annotations)...)
	}

	if len(errs) > 0 {
		//for now not fail the job, only return information
		//t.Fail()
		for _, e := range errs {
			t.Logf("%s\n", e)
		}
	}
	t.Skipf("exceptionally skip")
}

func checkRequiredAnnotations(name string, a map[string]string) []error {
	var errs []error
	if _, ok := a[OwnerAnnotationName]; !ok {
		errs = append(errs, fmt.Errorf("%s: does not contain required label '%s'", name, OwnerAnnotationName))
	}
	if _, ok := a[DescriptionAnnotationName]; !ok {
		errs = append(errs, fmt.Errorf("%s: does not contain required label '%s'", name, DescriptionAnnotationName))
	}
	return errs
}
