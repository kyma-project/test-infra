package prow

import (
	"errors"
	"fmt"

	"github.com/google/go-github/v48/github"
	"github.com/sirupsen/logrus"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

const (
	OrgDefaultClonePath        = "/home/prow/go/src/github.com"
	ProwConfigDefaultClonePath = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/config.yaml"
	JobConfigDefaultClonePath  = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/jobs"
	OwnerAnnotationName        = "owner"
	DescriptionAnnotationName  = "description"
)

// NotPresubmitError provides way to inform caller that prowjob is not a presubmit type.
// It provides brief error messages about it.
type NotPresubmitError struct{}

type MissingRequiredAnnotations struct {
	Owner       bool
	Description bool
	PjName      string
}

// non-colorable String() is used by go's string formatting support but ignored by ReportEntry
func (s MissingRequiredAnnotations) String() string {
	if s.Owner && s.Description {
		return fmt.Sprintf("[PJ config test failed] Prowjob %s is missing required annotations: %s, %s", s.PjName, OwnerAnnotationName, DescriptionAnnotationName)
	} else if s.Owner {
		return fmt.Sprintf("[PJ config test failed] Prowjob %s is missing required annotation: %s", s.PjName, OwnerAnnotationName)
	} else if s.Description {
		return fmt.Sprintf("[PJ config test failed] Prowjob %s is missing reuired annotation: %s", s.PjName, DescriptionAnnotationName)
	}
	return ""
}

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

// GetOrgForPresubmit will provide organization name for prowjob of type presubmit.
func GetOrgForPresubmit() (string, error) {
	// Get data, about prowjob specification from prowjob environment variables set by prow.
	jobSpec, err := downwardapi.ResolveSpecFromEnv()
	if err != nil {
		return "", fmt.Errorf("failed to read JOB_SPEC env, got error: %w", err)
	}
	// Get org name for presubmits only.
	if jobSpec.Type == prowapi.PresubmitJob {
		return jobSpec.Refs.Org, nil
	}
	return "", &NotPresubmitError{}
}

// GetProwjobsConfigForProwjob will provide all prowjobs configuration stored in repository described by orgName and repoName.
// If orgName and repoName are set to "kyma-project" and "test-infra" respectively, then static prowjobs configuration will be returned.
// Otherwise, inrepo prowjobs configuration will be returned.
// Prowjobs configuration is read from local repository clone.
func GetProwjobsConfigForProwjob(orgName, repoName, prowConfigPath, staticJobConfigPath, inrepoConfigPath string) ([]config.Presubmit, []config.Postsubmit, []config.Periodic, error) {
	var (
		presubmits  []config.Presubmit
		postsubmits []config.Postsubmit
		periodics   []config.Periodic
	)
	repoIdentifier := orgName + "/" + repoName
	conf, err := config.Load(prowConfigPath, staticJobConfigPath, nil, "")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading prow configs, got error: %w", err)
	}
	if orgName == "kyma-project" && repoName == "test-infra" {
		presubmits = conf.AllStaticPresubmits([]string{})
		postsubmits = conf.AllStaticPostsubmits([]string{})
		periodics = conf.AllPeriodics()
	} else {
		prowYAML, err := config.ReadProwYAML(logrus.WithField("repo", repoIdentifier), inrepoConfigPath, false)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error loading inrepo config, got error: %w", err)
		}
		presubmits = append(presubmits, prowYAML.Presubmits...)
		postsubmits = append(postsubmits, prowYAML.Postsubmits...)
	}
	return presubmits, postsubmits, periodics, nil
}

func CheckRequiredAnnotations(name string, a map[string]string) MissingRequiredAnnotations {
	var missingAnnotations MissingRequiredAnnotations
	if _, ok := a[OwnerAnnotationName]; !ok {
		missingAnnotations.Owner = true
	}
	if _, ok := a[DescriptionAnnotationName]; !ok {
		missingAnnotations.Description = true
	}
	if missingAnnotations.Owner || missingAnnotations.Description {
		missingAnnotations.PjName = name
		return missingAnnotations
	}
	return MissingRequiredAnnotations{}
}
