package diskscollector

import (
	"regexp"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"

	compute "google.golang.org/api/compute/v1"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=ZoneAPI -output=automock -outpkg=automock -case=underscore

// ZoneAPI allows to acces Zones Compute API in GCP
type ZoneAPI interface {
	ListZones(project string) ([]string, error)
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=DiskAPI -output=automock -outpkg=automock -case=underscore

// DiskAPI allows to access Disks Compute API in GCP
type DiskAPI interface {
	ListDisks(project, zone string) ([]*compute.Disk, error)
	RemoveDisk(name, project, zone string) error
}

// DisksGarbageCollector can find and delete disks provisioned by gke-integration prow jobs and not cleaned up properly.
type DisksGarbageCollector struct {
	zoneAPI      ZoneAPI
	diskAPI      DiskAPI
	shouldRemove DiskRemovalPredicate
}

// NewDisksGarbageCollector returns a new instance of DisksGarbageCollector
func NewDisksGarbageCollector(zoneAPI ZoneAPI, diskAPI DiskAPI, shouldRemove DiskRemovalPredicate) *DisksGarbageCollector {
	return &DisksGarbageCollector{zoneAPI, diskAPI, shouldRemove}
}

// Run executes disks garbage collection process
func (gc *DisksGarbageCollector) Run(project string, makeChanges bool) (allSucceeded bool, err error) {

	common.Shout("Looking for matching disks in \"%s\" project...", project)

	garbageDisks, err := gc.list(project)
	if err != nil {
		return
	}

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	if len(garbageDisks) > 0 {
		log.Infof("%sFound %d matching disks", msgPrefix, len(garbageDisks))
		common.Shout("Removing matching disks...")
	} else {
		log.Infof("%sFound no disks to delete", msgPrefix)
	}

	allSucceeded = true
	for _, gd := range garbageDisks {

		var err error

		if makeChanges {
			err = gc.diskAPI.RemoveDisk(project, gd.zone, gd.disk.Name)
		}

		if err != nil {
			log.Errorf("deleting disk %s: %#v", gd.disk.Name, err)
			allSucceeded = false
		} else {
			log.Infof("%sRequested disk delete: \"%s\". Project \"%s\", zone \"%s\", disk creationTimestamp: \"%s\"", msgPrefix, gd.disk.Name, project, gd.zone, gd.disk.CreationTimestamp)
		}
	}

	return allSucceeded, nil
}

type garbageDisk struct {
	zone string
	disk *compute.Disk
}

// list returns a filtered list of all disks in the project
// The list contains only disks that match removal criteria
func (gc *DisksGarbageCollector) list(project string) ([]*garbageDisk, error) {
	zones, err := gc.zoneAPI.ListZones(project)

	if err != nil {
		return nil, err
	}

	toRemove := []*garbageDisk{}

	for _, zone := range zones {
		disks, err := gc.diskAPI.ListDisks(project, zone)
		if err != nil {
			log.Errorf("listing disks for zone \"%s\": %#v", zone, err)
		}

		for _, disk := range disks {
			shouldRemove, err := gc.shouldRemove(disk)
			if err != nil {
				log.Warnf("Cannot check status of the disk %s due to an error: %#v", disk.Name, err)
			} else if shouldRemove {
				toRemove = append(toRemove, &garbageDisk{zone, disk})
			}
		}
	}

	return toRemove, nil
}

// DiskRemovalPredicate returns true when disk should be deleted (matches removal criteria)
type DiskRemovalPredicate func(*compute.Disk) (bool, error)

// NewDiskFilter is a default DiskRemovalPredicate factory
// Disk is matching the criteria if it's:
// - Name matches diskNameRegexp
// - CreationTimestamp indicates that it is created more than ageInHours ago.
// - Users list is empty
func NewDiskFilter(diskNameRegexp *regexp.Regexp, ageInHours int) DiskRemovalPredicate {
	return func(disk *compute.Disk) (bool, error) {
		nameMatches := diskNameRegexp.MatchString(disk.Name)
		useCountIsZero := len(disk.Users) == 0
		oldEnough := false

		diskCreationTime, err := time.Parse(time.RFC3339, disk.CreationTimestamp)
		if err != nil {
			log.Errorf("Error while parsing CreationTimestamp: \"%s\" for the disk: %s", disk.CreationTimestamp, disk.Name)
			return false, err
		}

		diskAgeHours := time.Since(diskCreationTime).Hours() - float64(ageInHours)
		oldEnough = diskAgeHours > 0

		if nameMatches && useCountIsZero && oldEnough {
			return true, nil
		}

		if nameMatches && oldEnough {
			message := "Found a disk that could be deleted but's still in use. Name: %s, age: %f[hours], use count: %d"
			log.Infof(message, disk.Name, diskAgeHours, len(disk.Users))
		}

		return false, nil
	}
}
