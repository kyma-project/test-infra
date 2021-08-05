package ipcleaner

import (
	"regexp"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"

	compute "google.golang.org/api/compute/v1"
)

//go:generate mockery --name=RegionAPI --output=automock --outpkg=automock --case=underscore

// RegionAPI abstracts access to Regions Compute API in GCP
type RegionAPI interface {
	ListRegions(project string) ([]string, error)
}

//go:generate mockery --name=AddressAPI --output=automock --outpkg=automock --case=underscore

// AddressAPI abstracts access to Address Compute API in GCP
type AddressAPI interface {
	ListAddresses(project, region string) ([]*compute.Address, error)
	RemoveIP(project, region, name string) error
}

// IPRemover deletes IPs provisioned by prow jobs.
type IPRemover struct {
	addressAPI   AddressAPI
	regionAPI    RegionAPI
	shouldRemove IPRemovalPredicate
}

// New returns a new instance of IPRemover
func New(addressAPI AddressAPI, regionAPI RegionAPI, shouldRemove IPRemovalPredicate) *IPRemover {
	return &IPRemover{addressAPI, regionAPI, shouldRemove}
}

type garbageAddress struct {
	region  string
	address *compute.Address
}

// Run executes ip removal process for specified IP
func (ipr *IPRemover) Run(project string, makeChanges bool) (allSucceeded bool, err error) {

	common.Shout("Looking for matching addresses in \"%s\" project...", project)

	garbageAddresses, err := ipr.list(project)
	if err != nil {
		return
	}

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	if len(garbageAddresses) > 0 {
		log.Infof("%sFound %d matching addresses", msgPrefix, len(garbageAddresses))
		common.Shout("Removing matching addresses...")
	} else {
		log.Infof("%sFound no addresses to delete", msgPrefix)
	}

	allSucceeded = true

	for _, ga := range garbageAddresses {
		var err error

		if makeChanges {
			err = ipr.addressAPI.RemoveIP(project, ga.region, ga.address.Name)
		}
		if err != nil {
			log.Errorf("deleting address %s: %#v", ga.address.Name, err)
			allSucceeded = false
		} else {
			log.Infof("%sRequested address delete: \"%s\". Project \"%s\", region \"%s\", address creationTimestamp: \"%s\"", msgPrefix, ga.address.Name, project, ga.region, ga.address.CreationTimestamp)
		}
	}

	return allSucceeded, nil
}

func (ipr *IPRemover) list(project string) ([]*garbageAddress, error) {
	regions, err := ipr.regionAPI.ListRegions(project)
	if err != nil {
		return nil, err
	}

	toRemove := []*garbageAddress{}

	for _, region := range regions {
		addresses, err := ipr.addressAPI.ListAddresses(project, region)
		if err != nil {
			log.Errorf("listing addresses for region \"%s\": %#v", region, err)
		}

		for _, address := range addresses {
			shouldRemove, err := ipr.shouldRemove(address)
			if err != nil {
				log.Warnf("Cannot check status of the address %s due to an error: %#v", address.Name, err)
			} else if shouldRemove {
				toRemove = append(toRemove, &garbageAddress{region, address})
			}
		}
	}

	return toRemove, nil
}

// IPRemovalPredicate returns true when address should be deleted (matches removal criteria)
type IPRemovalPredicate func(*compute.Address) (bool, error)

// NewIPFilter is a default IPRemovalPredicate factory
// Address is matching the criteria if it's:
// - Name does not match ipNameIgnoreRegexp
// - CreationTimestamp indicates that it is created more than ageInHours ago.
// - Users list is empty
func NewIPFilter(ipNameIgnoreRegexp *regexp.Regexp, ageInHours int) IPRemovalPredicate {
	return func(address *compute.Address) (bool, error) {
		nameMatches := ipNameIgnoreRegexp.MatchString(address.Name)
		useCountIsZero := len(address.Users) == 0
		oldEnough := false

		ipCreationTime, err := time.Parse(time.RFC3339, address.CreationTimestamp)
		if err != nil {
			log.Errorf("Error while parsing CreationTimestamp: \"%s\" for the ip: %s", address.CreationTimestamp, address.Name)
			return false, err
		}

		ipAgeHours := time.Since(ipCreationTime).Hours() - float64(ageInHours)
		oldEnough = ipAgeHours > 0

		if !nameMatches && useCountIsZero && oldEnough {
			return true, nil
		}

		if !nameMatches && oldEnough {
			message := "Found an IP that could be deleted but's still in use. Name: %s, age: %f[hours], use count: %d"
			log.Infof(message, address.Name, ipAgeHours, len(address.Users))
		}

		return false, nil
	}
}
