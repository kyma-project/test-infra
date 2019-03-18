package firewallcleaner

import (
	"context"

	log "github.com/sirupsen/logrus"

	compute "google.golang.org/api/compute/v1"
)

//ComputeServiceWrapper A wrapper for compute API service connections.
type ComputeServiceWrapper struct {
	Context context.Context
	Compute *compute.Service
}

//LookupFirewallRule List of all available firewall rules for a project
func (csw *ComputeServiceWrapper) LookupFirewallRule(project string) ([]*compute.Firewall, error) {
	call := csw.Compute.Firewalls.List(project)
	var items []*compute.Firewall
	f := func(page *compute.FirewallList) error {
		for _, list := range page.Items {
			items = append(items, list)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

//LookupGlobalForwardingRule ???
func (csw *ComputeServiceWrapper) LookupGlobalForwardingRule(project string) ([]*compute.ForwardingRule, error) {
	call := csw.Compute.GlobalForwardingRules.List(project)
	var items []*compute.ForwardingRule
	f := func(page *compute.ForwardingRuleList) error {
		for _, list := range page.Items {
			items = append(items, list)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

//LookupForwardingRule ???
func (csw *ComputeServiceWrapper) LookupForwardingRule(project, region string) ([]*compute.ForwardingRule, error) {
	call := csw.Compute.ForwardingRules.List(project, region)
	var items []*compute.ForwardingRule
	f := func(page *compute.ForwardingRuleList) error {
		for _, list := range page.Items {
			items = append(items, list)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

//LookupRegion ???
func (csw *ComputeServiceWrapper) LookupRegion(project string) ([]*compute.Region, error) {
	call := csw.Compute.Regions.List(project)
	var items []*compute.Region
	f := func(page *compute.RegionList) error {
		for _, list := range page.Items {
			items = append(items, list)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

//DeleteFirewallRule Delete firewall rule base on name in specifiec project
func (csw *ComputeServiceWrapper) DeleteFirewallRule(project, firewall string) {
	_, err := csw.Compute.Firewalls.Delete(project, firewall).Do()
	if err != nil {
		log.Print(err)
	}
}

//DeleteForwardingRule ???
func (csw *ComputeServiceWrapper) DeleteForwardingRule(project, name, region string) {
	_, err := csw.Compute.ForwardingRules.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

//DeleteGlobalForwardingRule ???
func (csw *ComputeServiceWrapper) DeleteGlobalForwardingRule(project, name string) {
	_, err := csw.Compute.GlobalForwardingRules.Delete(project, name).Do()
	if err != nil {
		log.Print(err)
	}
}
