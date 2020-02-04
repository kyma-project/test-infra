package clusterscollector

import (
	"errors"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"google.golang.org/api/container/v1"
)

const volatileLabelName = "volatile"
const createdAtLabelName = "created-at"
const ttlLabelName = "ttl"

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
			log.Errorf("Error deleting cluster. name: \"%s\", zone: \"%s\", error: %#v", cluster.Name, cluster.Zone, removeErr)
			allSucceeded = false
		} else {
			log.Infof("%sRequested cluster delete. name: \"%s\", zone \"%s\", createTime: \"%s\"", msgPrefix, cluster.Name, cluster.Zone, cluster.CreateTime)
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

// TimeBasedClusterRemovalPredicate returns an instance of ClusterRemovalPredicate that filters clusters based on label "volatile", "created-at", "ttl" and status
func TimeBasedClusterRemovalPredicate(whitelistedClusters map[string]struct{}) ClusterRemovalPredicate {
	return func(cluster *container.Cluster) (bool, error) {
		var timestamp int64
		var ageInHours uint64
		var err error
		if _, ok := whitelistedClusters[cluster.Name]; ok {
			log.Warnf("Cluster is whitelisted, deletion will be skipped. name: name: \"%s\", zone: \"%s\"", cluster.Name, cluster.Zone)
			return false, nil
		}

		isVolatileCluster := false
		if cluster.ResourceLabels != nil && cluster.ResourceLabels[volatileLabelName] == "true" {
			isVolatileCluster = true
		}

		isPastAgeMark := false
		if cluster.ResourceLabels != nil && cluster.ResourceLabels[createdAtLabelName] != "" {
			timestamp, err = strconv.ParseInt(cluster.ResourceLabels[createdAtLabelName], 10, 64)
			if err != nil {
				return false, errors.New("invalid timestamp value")
			}
		}

		if timestamp == 0 { // old cluster, does not have timestamp label, skip
			log.Warnf("Cluster does not have 'created-at' label, skipping. name: name: \"%s\", zone: \"%s\"", cluster.Name, cluster.Zone)
			return false, nil
		}

		hasTTL := false
		if cluster.ResourceLabels != nil && cluster.ResourceLabels[ttlLabelName] != "" {
			ageInHours, err = strconv.ParseUint(cluster.ResourceLabels[ttlLabelName], 10, 64)
			if err != nil {
				return false, errors.New("invalid ttl value")
			}
			hasTTL = true
		}

		createdAtTime := time.Unix(timestamp, 0)

		oneHourAgo := time.Now().Add(time.Duration(-ageInHours) * time.Hour)
		if oneHourAgo.After(createdAtTime) {
			isPastAgeMark = true
		}
		if hasTTL && isVolatileCluster && isPastAgeMark {
			//Filter out clusters that are being deleted at this moment
			if cluster.Status == "STOPPING" {
				log.Warnf("Cluster is already in STOPPING status, skipping. name: \"%s\", zone: \"%s\", createTime: \"%s\", label createdAt: \"%s\", ttl: \"%s\"", cluster.Name, cluster.Zone, cluster.CreateTime, cluster.ResourceLabels[createdAtLabelName], cluster.ResourceLabels[ttlLabelName])
				return false, nil
			}
			return true, nil
		}

		return false, nil
	}
}
