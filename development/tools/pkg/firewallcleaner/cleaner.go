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

//go:generate mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

//ComputeAPI interface logic for Google cloud API
type ComputeAPI interface {
	LookupFirewallRule(project string) ([]*compute.Firewall, error)
	LookupInstances(project string) ([]*compute.Instance, error)
	LookupNodePools(project string) ([]*container.NodePool, error)
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
	if err := c.checkAndDeleteFirewallRules(project, dryRun); err != nil {
		return errors.Wrap(err, fmt.Sprintf("checkAndDeleteFirewallRules failed for project '%s'", project))
	}
	common.Shout("Cleaner ran without errors")

	return nil
}

func (c *Cleaner) checkAndDeleteFirewallRules(project string, dryRun bool) error {
	rules, err := c.computeAPI.LookupFirewallRule(project)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("call to LookupFirewallRule failed for project '%s'", project))
	}
	instances, err := c.computeAPI.LookupInstances(project)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("call to LookupInstances failed for project '%s'", project))
	}

	nodePools, err := c.computeAPI.LookupNodePools(project)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("call to LookupNodePools failed for project '%s'", project))
	}
	poolNames := []string{}
	for _, pool := range nodePools {
		// group names are based on a cutoff of the cluster name at 24 characters (or full name if shorter), followed by -default-pool-[a-z0-9]+-grp
		// they are the last part of an instance group url
		if pool.InitialNodeCount > 0 && len(pool.InstanceGroupUrls) > 0 {
			str := pool.InstanceGroupUrls[0]
			split := strings.Split(str, "/")
			str = split[len(split)-1]

			regMatch := regexp.MustCompile("(.*)-default-pool-[a-z0-9]+-grp")

			matched := regMatch.FindStringSubmatch(str)
			if len(matched) > 1 {
				poolNames = append(poolNames, matched[1])
			}
		}
	}

	common.Shout("Collected %d rules, %d instances and %d node pools", len(rules), len(instances), len(nodePools))

	count := 0
	for _, rule := range rules {
		exist := false
		for _, target := range rule.TargetTags { // If no targetTags are specified, the firewall rule applies to all instances on the specified network. ref: https://cloud.google.com/compute/docs/reference/rest/v1/firewalls/list
			for _, instance := range instances {
				if instance.Name == target {
					exist = true
					break
				}
			}
		}
		for _, poolName := range poolNames {
			if strings.Contains(rule.Name, poolName) {
				exist = true
				break
			}
		}
		if strings.HasPrefix(rule.Name, "k8s-") {
			exist = true // ignore this rule
		}
		if !exist && len(rule.TargetTags) > 0 {
			count = count + 1
			if !dryRun {
				// c.computeAPI.DeleteFirewallRule(project, rule.Name)
				common.Shout("Deleting rule '%s' because there's no target running (%d TargetTags)", rule.Name, len(rule.TargetTags))
				time.Sleep(sleepFactor * time.Second)
			} else {
				common.Shout("[DRY RUN] Deleting rule '%s' because there's no target running (%d TargetTags)", rule.Name, len(rule.TargetTags))
			}
		}
	}
	common.Shout("Checked %d rules, deleted %d", len(rules), count)
	return nil
}
