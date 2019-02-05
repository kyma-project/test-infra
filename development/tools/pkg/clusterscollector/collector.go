package clusterscollector

import (
	"errors"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"google.golang.org/api/container/v1"
)

const volatileLabelName = "volatile"

//go:generate mockery -name=ClusterAPI -output=automock -outpkg=automock -case=underscore

// ClusterAPI abstracts access to Cluster Container API in GCP
type ClusterAPI interface {
	ListClusters(project string) ([]*container.Cluster, error)
	RemoveCluster(project, zone, name string) error
}

// ClustersGarbageCollector can find and delete clusters provisioned by gke-integration prow jobs and not cleaned up properly.
type ClustersGarbageCollector struct {
	clusterAPI   ClusterAPI
	shouldRemove ClusterRemovalPredicate
}

// NewClustersGarbageCollector returns a new instance of ClustersksGarbageCollector
func NewClustersGarbageCollector(clusterAPI ClusterAPI, shouldRemove ClusterRemovalPredicate) *ClustersGarbageCollector {
	return &ClustersGarbageCollector{clusterAPI, shouldRemove}
}

// Run executes garbage collection process for clusters
func (gc *ClustersGarbageCollector) Run(project string, makeChanges bool) (allSucceeded bool, err error) {

	common.Shout("Looking for matching clusters in \"%s\" project...", project)
	garbageClusters, err := gc.list(project)
	if err != nil {
		return
	}

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	if len(garbageClusters) > 0 {
		log.Infof("%sFound %d matching clusters", msgPrefix, len(garbageClusters))
		common.Shout("Removing matching clusters...")
	} else {
		log.Infof("%sFound no clusters to delete", msgPrefix)
	}

	allSucceeded = true
	for _, cluster := range garbageClusters {

		var removeErr error

		if makeChanges {
			removeErr = gc.clusterAPI.RemoveCluster(project, cluster.Zone, cluster.Name)
		}

		if removeErr != nil {
			log.Errorf("Error deleting cluster. Name: \"%s\", zone: \"%s\", error: %#v", cluster.Name, cluster.Zone, removeErr)
			allSucceeded = false
		} else {
			log.Infof("%sRequested cluster delete. Name: \"%s\", zone \"%s\", createTime: \"%s\"", msgPrefix, cluster.Name, cluster.Zone, cluster.CreateTime)
		}
	}

	return
}

// list returns a filtered list of all clusters in the project
// The list contains only clusters that match removal criteria
func (gc *ClustersGarbageCollector) list(project string) ([]*container.Cluster, error) {

	toRemove := []*container.Cluster{}

	clusters, err := gc.clusterAPI.ListClusters(project)
	if err != nil {
		log.Errorf("Error listing clusters : %#v", err)
		return nil, err
	}

	for _, cluster := range clusters {
		shouldRemove, err := gc.shouldRemove(cluster)
		if err != nil {
			log.Warnf("Cannot check status of the cluster %s due to an error: %#v", cluster.Name, err)
		} else if shouldRemove {
			toRemove = append(toRemove, cluster)
		}
	}

	return toRemove, nil
}

// ClusterRemovalPredicate returns true when the cluster should be deleted (matches removal criteria)
type ClusterRemovalPredicate func(cluster *container.Cluster) (bool, error)

// DefaultClusterRemovalPredicate returns an instance of ClusterRemovalPredicate that filters clusters based on clusterNameRegexp, label "volatile", ageInHours and Status
func DefaultClusterRemovalPredicate(clusterNameRegexp *regexp.Regexp, ageInHours uint) ClusterRemovalPredicate {
	return func(cluster *container.Cluster) (bool, error) {
		if cluster == nil {
			return false, errors.New("Invalid data: Nil")
		}

		nameMatches := clusterNameRegexp.MatchString(cluster.Name)

		isVolatileCluster := false
		if cluster.ResourceLabels != nil && cluster.ResourceLabels[volatileLabelName] == "true" {
			isVolatileCluster = true
		}

		var ageMatches bool

		clusterCreationTime, err := time.Parse(time.RFC3339, cluster.CreateTime)
		if err != nil {
			log.Errorf("Error while parsing CreateTime: \"%s\" for the cluster: %s", cluster.CreateTime, cluster.Name)
			return false, err
		}

		clusterAgeThreshold := time.Since(clusterCreationTime).Hours() - float64(ageInHours)
		ageMatches = clusterAgeThreshold > 0

		if nameMatches && isVolatileCluster && ageMatches {
			//Filter out clusters that are being deleted at this moment
			if cluster.Status == "STOPPING" {
				log.Warnf("Cluster is already in STOPPING status, skipping. Name: \"%s\", zone: \"%s\", createTime: \"%s\"", cluster.Name, cluster.Zone, cluster.CreateTime)
				return false, nil
			}
			return true, nil
		}

		return false, nil
	}
}
