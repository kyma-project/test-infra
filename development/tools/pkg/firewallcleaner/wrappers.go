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

//ListFirewallRules List of all available firewall rules for a project
func (csw *ComputeServiceWrapper) ListFirewallRules(project string) ([]*compute.Firewall, error) {
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

//DeleteFirewallRule Delete firewall rule base on name in specifiec project
func (csw *ComputeServiceWrapper) DeleteFirewallRule(project, firewall string) {
	_, err := csw.Compute.Firewalls.Delete(project, firewall).Do()
	if err != nil {
		log.Print(err)
	}
}
