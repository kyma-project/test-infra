package prow

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/google/go-github/v40/github"
	"github.com/sirupsen/logrus"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
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

func GetRepoProwjobsConfigForProwjob() ([]config.Presubmit, []config.Postsubmit, []config.Periodic, error) {
	orgName := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")
	repoIdentifier := orgName + "/" + repoName
	dir, err := os.Getwd()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get current working directory")
	}
	if orgName == "kyma-project" && repoName == "test-infra" {
		configPath := path.Join(dir, "config.yaml")
		jobConfigPath := path.Join(dir, "prow/jobs")
		conf, err := config.Load(configPath, jobConfigPath, nil, "")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error loading prow config: %w", err)
		}
		return conf.GetPresubmitsStatic(repoIdentifier), conf.GetPostsubmitsStatic(repoIdentifier), conf.AllPeriodics(), nil
	} else {
		prowYAML, err := config.ReadProwYAML(logrus.WithField("repo", repoIdentifier), dir, false)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error loading prow config: %w", err)
		}
		return prowYAML.Presubmits, prowYAML.Postsubmits, nil, nil
	}
}
