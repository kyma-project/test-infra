package networkscollector

import (
	"context"

	compute "google.golang.org/api/compute/v1"
)

// NetworkAPIWrapper abstracts GCP NetworksService API
type NetworkAPIWrapper struct {
	Context context.Context
	Service *compute.NetworksService
}

// ListNetworks delegates to NetworksService.List(parent) function
func (naw *NetworkAPIWrapper) ListNetworks(project string) ([]*compute.Network, error) {
	nl, err := naw.Service.List(project).Context(naw.Context).Do()
	if err != nil {
		return nil, err
	}

	return nl.Items, nil
}

// RemoveNetwork delegates to NetworksService.Delete(name) function
func (naw *NetworkAPIWrapper) RemoveNetwork(project, name string) error {
	_, err := naw.Service.Delete(project, name).Do()

	if err != nil {
		return err
	}

	return nil
}
