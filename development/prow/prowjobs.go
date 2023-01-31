package prow

import (
	"errors"
	"fmt"

	"github.com/google/go-github/v40/github"
	"github.com/sirupsen/logrus"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

const (
	OrgDefaultClonePath        = "/home/prow/go/src/github.com"
	TestInfraDefaultClonePath  = "/home/prow/go/src/github.com/kyma-project/test-infra"
	ProwConfigDefaultClonePath = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/config.yaml"
	JobConfigDefaultClonePath  = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/jobs"
)

// NotPresubmitError provides way to inform caller that prowjob is not a presubmit type.
// It provides brief error messages about it.
type NotPresubmitError struct{}

func (e *NotPresubmitError) Error() string { return "prowjob is not a presubmit type" }

// IsNotPresubmitError checks if error is of type NotPresubmitError.
func IsNotPresubmitError(err error) *bool {
	var e *NotPresubmitError
	if errors.As(err, &e) {
		return github.Bool(true)
	}
	return github.Bool(false)

}

// GetPrAuthorForPresubmit will provide list all pull requests authors for prowjob of type presubmit.
// Use IsNotPresbumitError to check if GetPrAuthorForPresubmit was called against other types of prowjobs.
func GetPrAuthorForPresubmit() ([]string, error) {
	// Get data, about prowjob specification from prowjob environment variables set by prow.
	jobSpec, err := downwardapi.ResolveSpecFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to read JOB_SPEC env, got error: %w", err)
	}
	// Get authors for presubmits only.
	if jobSpec.Type == prowapi.PresubmitJob {
		var authors []string
		// Go through all presubmit pulls to get authors.
		for _, pull := range jobSpec.Refs.Pulls {
			authors = append(authors, pull.Author)
		}
		return authors, nil
	}
	return nil, &NotPresubmitError{}
}

func GetOrgForPresubmit() (string, error) {
	// Get data, about prowjob specification from prowjob environment variables set by prow.
	jobSpec, err := downwardapi.ResolveSpecFromEnv()
	if err != nil {
		return "", fmt.Errorf("failed to read JOB_SPEC env, got error: %w", err)
	}
	// Get authors for presubmits only.
	if jobSpec.Type == prowapi.PresubmitJob {
		return jobSpec.Refs.Org, nil
	}
	return "", &NotPresubmitError{}
}

func GetProwjobsConfigForProwjob(orgName, repoName, prowConfigPath, staticJobConfigPath, inrepoConfigPath string) ([]config.Presubmit, []config.Postsubmit, []config.Periodic, error) {
	repoIdentifier := orgName + "/" + repoName
	conf, err := config.Load(prowConfigPath, staticJobConfigPath, nil, "")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading prow configs, got error: %w", err)
	}
	presubmits := conf.GetPresubmitsStatic(repoIdentifier)
	postsubmits := conf.GetPostsubmitsStatic(repoIdentifier)
	periodics := conf.AllPeriodics()
	if orgName != "kyma-project" && repoName != "test-infra" {
		prowYAML, err := config.ReadProwYAML(logrus.WithField("repo", repoIdentifier), inrepoConfigPath, false)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error loading inrepo config, got error: %w", err)
		}
		presubmits = append(presubmits, prowYAML.Presubmits...)
		postsubmits = append(postsubmits, prowYAML.Postsubmits...)
	}
	return presubmits, postsubmits, periodics, nil
}
