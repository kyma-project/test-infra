package networkscollector

import (
	"errors"
	"regexp"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	compute "google.golang.org/api/compute/v1"
)

//go:generate mockery -name=NetworkAPI -output=automock -outpkg=automock -case=underscore

// NetworkAPI abstracts access to Network Compute API in GCP
type NetworkAPI interface {
	ListNetworks(project string) ([]*compute.Network, error)
	RemoveNetwork(project, name string) error
}

// NetworksGarbageCollector can find and delete networks provisioned by gke-integration prow jobs and not cleaned up properly.
type NetworksGarbageCollector struct {
	networkAPI   NetworkAPI
	shouldRemove NetworkRemovalPredicate
}

// NewNetworksGarbageCollector returns a new instance of NetworksGarbageCollector
func NewNetworksGarbageCollector(networkAPI NetworkAPI, shouldRemove NetworkRemovalPredicate) *NetworksGarbageCollector {
	return &NetworksGarbageCollector{networkAPI, shouldRemove}
}

// Run executes garbage collection process for networks
func (gc *NetworksGarbageCollector) Run(project string, makeChanges bool) (allSucceeded bool, err error) {

	common.Shout("Looking for matching networks in \"%s\" project...", project)
	garbageNetworks, err := gc.list(project)
	if err != nil {
		return
	}

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	if len(garbageNetworks) > 0 {
		log.Infof("%sFound %d matching networks", msgPrefix, len(garbageNetworks))
		common.Shout("Removing matching networks...")
	} else {
		log.Infof("%sFound no networks to delete", msgPrefix)
	}

	allSucceeded = true
	for _, network := range garbageNetworks {

		var removeErr error

		if makeChanges {
			removeErr = gc.networkAPI.RemoveNetwork(project, network.Name)
		}

		if removeErr != nil {
			log.Errorf("Error deleting network. Name: \"%s\", error: %#v", network.Name, removeErr)
			allSucceeded = false
		} else {
			log.Infof("%sRequested network delete. Name: \"%s\", createTime: \"%s\"", msgPrefix, network.Name, network.CreationTimestamp)
		}
	}

	return
}

// list returns a filtered list of all networks in the project
// The list contains only networks that match removal criteria
func (gc *NetworksGarbageCollector) list(project string) ([]*compute.Network, error) {

	toRemove := []*compute.Network{}

	networks, err := gc.networkAPI.ListNetworks(project)
	if err != nil {
		log.Errorf("Error listing networks : %#v", err)
		return nil, err
	}

	for _, network := range networks {
		shouldRemove, err := gc.shouldRemove(network)
		if err != nil {
			log.Warnf("Cannot check status of the network %s due to an error: %#v", network.Name, err)
		} else if shouldRemove {
			toRemove = append(toRemove, network)
		}
	}

	return toRemove, nil
}

// NetworkRemovalPredicate returns true when the network should be deleted (matches removal criteria)
type NetworkRemovalPredicate func(network *compute.Network) (bool, error)

// DefaultNetworkRemovalPredicate returns an instance of NetworkRemovalPredicate that filters networks based on ageInHours and Status
func DefaultNetworkRemovalPredicate(networkNameRegexp *regexp.Regexp, ageInHours uint) NetworkRemovalPredicate {
	return func(network *compute.Network) (bool, error) {
		if network == nil {
			return false, errors.New("Invalid data: Nil")
		}

		nameMatches := networkNameRegexp.MatchString(network.Name)

		var ageMatches bool

		networkCreationTime, err := time.Parse(time.RFC3339, network.CreationTimestamp)
		if err != nil {
			log.Errorf("Error while parsing CreateTime: \"%s\" for the cluster: %s", network.CreationTimestamp, network.Name)
			return false, err
		}

		ageThreshold := time.Since(networkCreationTime).Hours() - float64(ageInHours)
		ageMatches = ageThreshold > 0

		if nameMatches && ageMatches {
			return true, nil
		}

		return false, nil
	}
}
