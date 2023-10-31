package firewallcleaner

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
)

// ComputeServiceWrapper A wrapper for compute API service connections.
type ComputeServiceWrapper struct {
	Context   context.Context
	Compute   *compute.Service
	Container *container.Service
}

// LookupFirewallRule List of all available firewall rules for a project
func (csw *ComputeServiceWrapper) LookupFirewallRule(project string) ([]*compute.Firewall, error) {
	call := csw.Compute.Firewalls.List(project)
	var items []*compute.Firewall
	f := func(page *compute.FirewallList) error {
		items = append(items, page.Items...)
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, errors.Wrap(err, "LookupFirewallRule page switch failed")
	}
	return items, nil
}

// LookupInstances ???
func (csw *ComputeServiceWrapper) LookupInstances(project string) ([]*compute.Instance, error) {
	call := csw.Compute.Instances.AggregatedList(project)
	var items []*compute.Instance
	f := func(page *compute.InstanceAggregatedList) error {
		for _, list := range page.Items {
			items = append(items, list.Instances...)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, errors.Wrap(err, "LookupInstances page switch failed")
	}
	return items, nil
}

// LookupNodePools ???
func (csw *ComputeServiceWrapper) LookupNodePools(clusters []*container.Cluster) ([]*container.NodePool, error) {

	allClustersPools := []*container.NodePool{}
	for _, cluster := range clusters {
		allClustersPools = append(allClustersPools, cluster.NodePools...)
	}
	return allClustersPools, nil
}

// LookupClusters ???
func (csw *ComputeServiceWrapper) LookupClusters(project string) ([]*container.Cluster, error) {
	resp, err := csw.Container.Projects.Zones.Clusters.List(project, "-").Do() // "-" will get clusters in all zones
	if err != nil {
		return nil, err
	}
	return resp.Clusters, nil
}

// DeleteFirewallRule Delete firewall rule base on name in specifiec project
func (csw *ComputeServiceWrapper) DeleteFirewallRule(project, firewall string) {
	_, err := csw.Compute.Firewalls.Delete(project, firewall).Do()
	if err != nil {
		log.Print(errors.Wrap(err, "DeleteFirewallRule failed"))
	}
}
