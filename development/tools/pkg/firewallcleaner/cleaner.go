package firewallcleaner

import (
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	compute "google.golang.org/api/compute/v1"
)

const sleepFactor = 2

//go:generate mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

//ComputeAPI interface logic for Google cloud API
type ComputeAPI interface {
	LookupFirewallRule(project string) ([]*compute.Firewall, error)
	LookupGlobalForwardingRule(project string) ([]*compute.ForwardingRule, error)
	LookupForwardingRule(project, region string) ([]*compute.ForwardingRule, error)
	LookupRegion(project string) ([]*compute.Region, error)
	DeleteFirewallRule(project, firewall string)
	DeleteForwardingRule(project, name, region string)
	DeleteGlobalForwardingRule(project, name string)
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
	if err := c.checkAndDeleteFirewallRules(project); err != nil {
		return err
	}

	if err := c.checkAndDeleteGlobalForwardingRules(project); err != nil {
		return err
	}

	if err := c.checkAndDeleteForwardingRules(project); err != nil {
		return err
	}

	return nil
}

func (c *Cleaner) checkAndDeleteFirewallRules(project string) error {
	rules, err := c.computeAPI.LookupFirewallRule(project)
	if err != nil {
		return err
	}
	for _, r := range rules {
		for _, t := range r.TargetTags {
			if strings.Contains(t, "test") {
				// c.computeAPI.DeleteFirewallRule(project, r.Name)
				common.Shout("If I were serious, I'd delete the rule for the above PR here, because of the target tag. Rule name: %s, TargetTag: %s", r.Name, t)
			}
		}
	}
	return nil
}

func (c *Cleaner) checkAndDeleteGlobalForwardingRules(project string) error {
	rules, err := c.computeAPI.LookupGlobalForwardingRule(project)
	if err != nil {
		return err
	}
	for _, r := range rules {
		if strings.Contains(r.Target, "test") {
			// c.computeAPI.DeleteGlobalForwardingRule(project, r.Name)
			common.Shout("If I were serious, I'd delete the rule for the above PR here, because of the target tag. Rule name: %s, TargetTag: %s", r.Name, r.Target)
		}
	}
	return nil
}

func (c *Cleaner) checkAndDeleteForwardingRules(project string) error {
	regions, err := c.computeAPI.LookupRegion(project)
	if err != nil {
		return err
	}
	for _, region := range regions {
		rules, err := c.computeAPI.LookupForwardingRule(project, region.Name)
		if err != nil {
			return err
		}
		for _, r := range rules {
			if strings.Contains(r.Target, "test") {
				// c.computeAPI.DeleteForwardingRule(project, r.Name)
				common.Shout("If I were serious, I'd delete the rule for the above PR here, because of the target tag. Rule name: %s, TargetTag: %s", r.Name, r.Target)
			}
		}
	}
	return nil
}
