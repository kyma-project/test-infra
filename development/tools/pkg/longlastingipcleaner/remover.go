package ipcleaner

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

// Run executes ip removal process for specified IP
func (ipr *IPRemover) Run(project, region, ipName string, maxAttempts, timeout uint, makeChanges bool) (bool, error) {

	common.Shout("Trying to delete IP with name \"%s\" in project \"%s\", available in region \"%s\"", ipName, project, region)

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	var err error
	succeeded := true
	attempts := uint(0)
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	if makeChanges {
		for {
			retryable, removalErr := ipr.computeAPI.RemoveIP(project, region, ipName)
			log.Infof("retryable: %v, attempts: %d, err: %v\n", retryable, attempts, removalErr)
			attempts = attempts + 1
			if attempts < maxAttempts && retryable {
				time.Sleep(time.Duration(timeout) * time.Second)
				timeout = timeout * 2
			} else {
				err = removalErr
				break
			}
		}
	}
	if err != nil {
		log.Infof("Could not delete IP with name \"%s\" in region \"%s\", got error: %s", ipName, region, err.Error())
		succeeded = false
	} else {
		log.Infof("%sRequested deletion of IP with name \"%s\" in region \"%s\"", msgPrefix, ipName, region)
	}

	return succeeded, err
}
