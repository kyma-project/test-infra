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
	LookupForwardingRule(project string) ([]*compute.ForwardingRule, error)
	LookupInstances(project string) ([]*compute.Instance, error)
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
	instances, err := c.computeAPI.LookupInstances(project)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		for _, target := range rule.TargetTags {
			exist := false
			for _, instance := range instances {
				if instance.Name == target {
					exist = true
				}
			}
			if !exist {
				// c.computeAPI.DeleteFirewallRule(project, r.Name)
				common.Shout("If I were serious, I'd delete the rule for the above PR here, because of the target tag. Rule name: %s, TargetTag: %s", rule.Name, target)
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
	instances, err := c.computeAPI.LookupInstances(project)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		exist := false
		for _, instance := range instances {
			if instance.Name == rule.Target {
				exist = true
			}
		}
		if !exist {
			// c.computeAPI.DeleteGlobalForwardingRule(project, r.Name)
			common.Shout("If I were serious, I'd delete the rule for the above PR here, because of the target tag. Rule name: %s, TargetTag: %s", rule.Name, rule.Target)
		}
	}
	return nil
}

func (c *Cleaner) checkAndDeleteForwardingRules(project string) error {
	rules, err := c.computeAPI.LookupForwardingRule(project)
	if err != nil {
		return err
	}
	instances, err := c.computeAPI.LookupInstances(project)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		exist := false
		for _, instance := range instances {
			if instance.Name == rule.Target && strings.Contains(instance.Zone, rule.Region) {
				exist = true
			}
		}
		if !exist {
			// c.computeAPI.DeleteForwardingRule(project, r.Name)
			common.Shout("If I were serious, I'd delete the rule for the above PR here, because of the target tag. Rule name: %s, TargetTag: %s", rule.Name, rule.Target)
		}
	}
	return nil
}
