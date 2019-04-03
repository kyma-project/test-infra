package firewallcleaner

import (
	"fmt"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
)

const sleepFactor = 2

//go:generate mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

//ComputeAPI interface logic for Google cloud API
type ComputeAPI interface {
	LookupFirewallRule(project string) ([]*compute.Firewall, error)
	LookupInstances(project string) ([]*compute.Instance, error)
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
	for _, rule := range rules {
		for _, target := range rule.TargetTags { // If no targetTags are specified, the firewall rule applies to all instances on the specified network. ref: https://cloud.google.com/compute/docs/reference/rest/v1/firewalls/list
			exist := false
			for _, instance := range instances {
				if instance.Name == target {
					exist = true
				}
			}
			if !exist {
				if !dryRun {
					// c.computeAPI.DeleteFirewallRule(project, rule.Name)
					common.Shout("Deleting rule '%s' because there's no target '%s' running (%d TargetTags)", rule.Name, target, len(rule.TargetTags))
					time.Sleep(sleepFactor * time.Second)
				} else {
					common.Shout("[DRY RUN] Deleting rule '%s' because there's no target '%s' running (%d TargetTags)", rule.Name, target, len(rule.TargetTags))
				}
			}
		}
	}
	return nil
}
