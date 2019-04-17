package ipcleaner

import (
	"context"

	compute "google.golang.org/api/compute/v1"
)

// ComputeAPIWrapper abstracts GCP Compute Service API
type ComputeAPIWrapper struct {
	Context context.Context
	Service *compute.Service
}

// RemoveIP delegates to Compute.Service.Addresses.Delete(project, region, name) function
func (caw *ComputeAPIWrapper) RemoveIP(project, region, name string) error {
	_, err := caw.Service.Addresses.Delete(project, region, name).Context(caw.Context).Do()
	if err != nil {
		return err
	}

	return nil
}
