package iprelease

import (
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

// ComputeAPI abstracts access to Compute API in GCP
type ComputeAPI interface {
	RemoveIP(project, region, name string) (bool, error)
}

// IPRemover deletes IPs provisioned by gke-long-lasting prow jobs.
type IPRemover struct {
	computeAPI ComputeAPI
}

// NewIPRemover returns a new instance of IPRemover
func NewIPRemover(computeAPI ComputeAPI) *IPRemover {
	return &IPRemover{computeAPI}
}

// Run executes garbage collection process for clusters
func (ipr *IPRemover) Run(project, ipName, region string, maxAttempts, timeout int, makeChanges bool) (allSucceeded bool, err error) {

	common.Shout("Trying to delete IP with name \"%s\" in project \"%s\", available in region \"%s\"", ipName, project, region)

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	allSucceeded = true
	retryable := true
	attempts := 0
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	if makeChanges {
		for {
			retryable, err = ipr.computeAPI.RemoveIP(project, region, ipName)
			attempts = attempts + 1
			if err != nil || !retryable || attempts >= maxAttempts {
				break
			} else {
				time.Sleep(time.Duration(timeout) * time.Second)
				timeout = timeout * 2
			}
		}
	}
	if err != nil {
		log.Infof("Could not delete IP with name \"%s\" in region \"%s\", got error: %s", ipName, region, err.Error())
		allSucceeded = false
	} else {
		log.Infof("%sRequested deletion of IP with name \"%s\" in region \"%s\"", msgPrefix, ipName, region)
	}

	return
}
