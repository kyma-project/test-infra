package ipcleaner

import (
	"context"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

// ComputeAPI abstracts access to Compute API in GCP
type ComputeAPI interface {
	RemoveIP(ctx context.Context, project, region, name string) error
}

// IPRemover deletes IPs provisioned by gke-long-lasting prow jobs.
type IPRemover struct {
	computeAPI  ComputeAPI
	maxAttempts uint
	backoff     uint
	makeChanges bool
}

// New returns a new instance of IPRemover
func New(computeAPI ComputeAPI, maxAttempts, backoff uint, makeChanges bool) *IPRemover {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &IPRemover{computeAPI, maxAttempts, backoff, makeChanges}
}

// Run executes ip removal process for specified IP
func (ipr *IPRemover) Run(project, region, ipName string) error {
	common.Shout("Trying to delete IP with name \"%s\" in project \"%s\", available in region \"%s\"", ipName, project, region)

	var msgPrefix string
	if !ipr.makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	ctx := context.Background()

	var err error
	backoff := ipr.backoff
	for attempt := uint(0); attempt < ipr.maxAttempts; attempt = attempt + 1 {
		var removalErr error
		if ipr.makeChanges {
			removalErr = ipr.computeAPI.RemoveIP(ctx, project, region, ipName)
		}
		if removalErr == nil {
			log.Errorf("%sRequested deletion of IP with name \"%s\" in region \"%s\"", msgPrefix, ipName, region)
			return nil
		}
		if removalErr.Error() == ipDeletionFailed {
			return errors.Wrap(removalErr, "unable to delete non-existant ip")
		}
		if attempt < ipr.maxAttempts {
			time.Sleep(time.Duration(backoff) * time.Second)
			backoff = backoff * 2
			err = removalErr
		}
	}
	if err != nil {
		log.Errorf("Could not delete IP with name \"%s\" in region \"%s\", got error: %s", ipName, region, err.Error())
	}
	return errors.Wrap(err, "unable to delete ip")
}
