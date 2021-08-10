package prow

import (
	"errors"
	"fmt"

	"github.com/google/go-github/github"
	prowjobs "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

type NotPresubmitError struct{}

func (e *NotPresubmitError) Error() string { return "prowjob is not a presubmit type" }

func IsNotPresubmitError(err error) *bool {
	var e *NotPresubmitError
	if errors.As(err, &e) {
		return github.Bool(true)
	} else {
		return github.Bool(false)
	}
}

// GetPrAuthorForPresubmit
// Use IsNotPresbumitError
func GetPrAuthorForPresubmit() ([]string, error) {
	// get git base reference from postsubmit environment variables
	jobSpec, err := downwardapi.ResolveSpecFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to read JOB_SPEC env, got error: %w", err)
	}
	if jobSpec.Type == prowjobs.PresubmitJob {
		var authors []string
		for _, pull := range jobSpec.Refs.Pulls {
			authors = append(authors, pull.Author)
		}
		return authors, nil
	} else {
		return nil, &NotPresubmitError{}
	}
}
