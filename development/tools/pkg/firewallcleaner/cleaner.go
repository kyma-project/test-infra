package firewallcleaner

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
)

const sleepFactor = 2

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

//ComputeAPI interface logic for Google cloud API
type ComputeAPI interface {
	LookupFirewallRule(project string) ([]*compute.Firewall, error)
	LookupInstances(project string) ([]*compute.Instance, error)
	LookupNodePools(clusters []*container.Cluster) ([]*container.NodePool, error)
	LookupClusters(project string) ([]*container.Cluster, error)
	DeleteFirewallRule(project, firewall string)
}

//Cleaner Element holding the firewall cleaning logic
type Cleaner struct {
	computeAPI ComputeAPI
}

//NewCleaner Returns a new cleaner object
func NewCleaner(computeAPI ComputeAPI) *Cleaner {
	return &Cleaner{computeAPI}
}

//Run the main find&destroy function
func (c *Cleaner) Run(dryRun bool, project string) error {
	dryRunPrefix := ""
	if dryRun {
		dryRunPrefix = "[DRY RUN] "
	}
	if err := c.checkAndDeleteFirewallRules(project, dryRun); err != nil {
		return errors.Wrap(err, fmt.Sprintf("checkAndDeleteFirewallRules failed for project '%s'", project))
	}
	common.Shout("%sCleaner ran without errors", dryRunPrefix)

	return nil
}

func (c *Cleaner) checkAndDeleteFirewallRules(project string, dryRun bool) error {
	dryRunPrefix := ""
	if dryRun {
		dryRunPrefix = "[DRY RUN] "
	}
	rules, firewallErr := c.computeAPI.LookupFirewallRule(project)
	if firewallErr != nil {
		return errors.Wrap(firewallErr, fmt.Sprintf("call to LookupFirewallRule failed for project '%s'", project))
	}
	instances, instanceErr := c.computeAPI.LookupInstances(project)
	if instanceErr != nil {
		return errors.Wrap(instanceErr, fmt.Sprintf("call to LookupInstances failed for project '%s'", project))
	}

	clusters, clusterErr := c.computeAPI.LookupClusters(project)
	if clusterErr != nil {
		return errors.Wrap(clusterErr, fmt.Sprintf("call to LookupClusters failed for project '%s'", project))
	}

	nodePools, nodePoolErr := c.computeAPI.LookupNodePools(clusters)
	if nodePoolErr != nil {
		return errors.Wrap(nodePoolErr, fmt.Sprintf("call to LookupNodePools failed for project '%s'", project))
	}
	poolNames := []string{}
	for _, pool := range nodePools {
		// group names are based on a cutoff of the cluster name at 24 characters (or full name if shorter), followed by -default-pool-[a-z0-9]+-grp
		// they are the last part of an instance group url
		if pool.InitialNodeCount > 0 && len(pool.InstanceGroupUrls) > 0 {
			str := pool.InstanceGroupUrls[0]
			split := strings.Split(str, "/")
			str = split[len(split)-1]

			regMatch := regexp.MustCompile("(.*)-default-pool-[a-z0-9]+")

			matched := regMatch.FindStringSubmatch(str)
			if len(matched) > 1 {
				poolNames = append(poolNames, matched[1])
			}
		}
	}

	common.Shout("%sCollected %d rules, %d instances and %d node pools", dryRunPrefix, len(rules), len(instances), len(nodePools))

	count := 0
	for _, rule := range rules {
		exist := false
		for _, target := range rule.TargetTags { // If no targetTags are specified, the firewall rule applies to all instances on the specified network. ref: https://cloud.google.com/compute/docs/reference/rest/v1/firewalls/list
			for _, instance := range instances {
				if instance.Name == target {
					exist = true
					continue
				}
			}
			for _, cluster := range clusters { // takes care of 'k8s-' rules
				if strings.HasPrefix(target, cluster.Name) {
					exist = true
					continue
				}
			}
			for _, poolName := range poolNames {
				if strings.HasPrefix(target, poolName) {
					exist = true
					continue
				}
			}
		}
		for _, poolName := range poolNames {
			if strings.Contains(rule.Name, poolName) {
				exist = true
				continue
			}
		}
		if !exist && len(rule.TargetTags) > 0 {
			count = count + 1
			if !dryRun {
				c.computeAPI.DeleteFirewallRule(project, rule.Name)
				common.Shout("Deleting rule '%s' because there's no target running (%d TargetTags: %v)", rule.Name, len(rule.TargetTags), rule.TargetTags)
				time.Sleep(sleepFactor * time.Second)
			} else {
				common.Shout("[DRY RUN] Deleting rule '%s' because there's no target running (%d TargetTags: %v)", rule.Name, len(rule.TargetTags), rule.TargetTags)
			}
		}
	}
	common.Shout("%sChecked %d rules, deleted %d", dryRunPrefix, len(rules), count)
	return nil
}
