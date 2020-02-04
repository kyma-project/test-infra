package vmscollector

import (
	"errors"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	compute "google.golang.org/api/compute/v1"
)

const jobLabelName = "job-name"

//go:generate mockery -name=InstancesAPI -output=automock -outpkg=automock -case=underscore

// InstancesAPI abstracts access to Instances Compute API in GCP
type InstancesAPI interface {
	ListInstances(project string) ([]*compute.Instance, error)
	RemoveInstance(project, zone, name string) error
}

// InstancesGarbageCollector can find and delete VM instances provisioned by kyma-integration prow jobs and not cleaned up properly.
type InstancesGarbageCollector struct {
	instancesAPI InstancesAPI
	shouldRemove InstanceRemovalPredicate
}

// NewInstancesGarbageCollector returns a new object of InstancesGarbageCollector type
func NewInstancesGarbageCollector(instancesAPI InstancesAPI, shouldRemove InstanceRemovalPredicate) *InstancesGarbageCollector {
	return &InstancesGarbageCollector{instancesAPI, shouldRemove}
}

// Run executes garbage collection process for VM instances
func (gc *InstancesGarbageCollector) Run(project string, makeChanges bool) (allSucceeded bool, err error) {

	common.Shout("Looking for matching instances in \"%s\" project...", project)
	garbageInstances, err := gc.list(project)
	if err != nil {
		return
	}

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	if len(garbageInstances) > 0 {
		log.Infof("%sFound %d matching VM instances", msgPrefix, len(garbageInstances))
		common.Shout("Removing matching VM instances...")
	} else {
		log.Infof("%sFound no VM instances to delete", msgPrefix)
	}

	allSucceeded = true
	for _, instance := range garbageInstances {

		var removeErr error

		if makeChanges {
			removeErr = gc.instancesAPI.RemoveInstance(project, formatZone(instance.Zone), instance.Name)
		}

		if removeErr != nil {
			log.Errorf("Error deleting VM instance. name: \"%s\", zone: \"%s\", error: %#v", instance.Name, formatZone(instance.Zone), removeErr)
			allSucceeded = false
		} else {
			log.Infof("%sRequested VM instance delete. name: \"%s\", zone \"%s\", creationTimestamp: \"%s\"", msgPrefix, instance.Name, formatZone(instance.Zone), instance.CreationTimestamp)
		}
	}

	return
}

// list returns a filtered list of all VM instances in the project
// The list contains only VM instances that match removal criteria
func (gc *InstancesGarbageCollector) list(project string) ([]*compute.Instance, error) {

	toRemove := []*compute.Instance{}

	instances, err := gc.instancesAPI.ListInstances(project)
	if err != nil {
		log.Errorf("Error listing VM instances : %#v", err)
		return nil, err
	}

	for _, instance := range instances {
		shouldRemove, err := gc.shouldRemove(instance)
		if err != nil {
			log.Warnf("Cannot check status of the VM instance %s due to an error: %#v", instance.Name, err)
		} else if shouldRemove {
			toRemove = append(toRemove, instance)
		}
	}

	return toRemove, nil
}

// InstanceRemovalPredicate returns true when the VM instance should be deleted (matches removal criteria)
type InstanceRemovalPredicate func(instance *compute.Instance) (bool, error)

// DefaultInstanceRemovalPredicate returns an instance of InstanceRemovalPredicate that filters instances based on instanceNameRegexp, jobLabelRegexp, ageInHours and Status
func DefaultInstanceRemovalPredicate(instanceNameRegexp *regexp.Regexp, jobLabelRegexp *regexp.Regexp, ageInHours uint) InstanceRemovalPredicate {
	return func(instance *compute.Instance) (bool, error) {
		if instance == nil {
			return false, errors.New("Invalid data: Nil")
		}

		nameMatches := instanceNameRegexp.MatchString(instance.Name)

		jobLabelMatches := false
		if instance.Labels != nil {
			jobLabelMatches = jobLabelRegexp.MatchString(instance.Labels[jobLabelName])
		}

		var ageMatches bool

		instanceCreationTime, err := time.Parse(time.RFC3339, instance.CreationTimestamp)
		if err != nil {
			log.Errorf("Error while parsing creationTimestamp: \"%s\" for the VM instance: %s", instance.CreationTimestamp, instance.Name)
			return false, err
		}

		instanceAgeThreshold := time.Since(instanceCreationTime).Hours() - float64(ageInHours)
		ageMatches = instanceAgeThreshold > 0

		if nameMatches && jobLabelMatches && ageMatches {
			//Filter out instances that are not RUNNING at this moment
			if instance.Status != "RUNNING" {
				log.Warnf("VM Instance is not in RUNNING status, skipping. name: \"%s\", zone: \"%s\", creationTimestamp: \"%s\", status: \"%s\"", instance.Name, formatZone(instance.Zone), instance.CreationTimestamp, instance.Status)
				return false, nil
			}
			return true, nil
		}

		return false, nil
	}
}

func formatZone(zone string) string {
	splits := strings.Split(zone, "/")
	return splits[len(splits)-1]
}
