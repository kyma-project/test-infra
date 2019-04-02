package iprelease

import (
	"context"
	"net/http"

	compute "google.golang.org/api/compute/v1"
)

// ComputeAPIWrapper abstracts GCP Compute Service API
type ComputeAPIWrapper struct {
	Context context.Context
	Service *compute.Service
}

// RemoveIP delegates to Service.Addresses.Delete(project, region, name) function
func (caw *ComputeAPIWrapper) RemoveIP(project, region, name string) (bool, error) {
	resp, err := caw.Service.Addresses.Delete(project, region, name).Do()
	if err != nil {
		return false, err
	}
	if resp.HTTPStatusCode == http.StatusTooManyRequests {
		return true, err
	}

	return false, nil
}
