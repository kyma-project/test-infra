package ipcleaner

import (
	"context"
	"errors"
	"net/http"

	compute "google.golang.org/api/compute/v1"
)

// ComputeAPIWrapper abstracts GCP Compute Service API
type ComputeAPIWrapper struct {
	Context context.Context
	Service *compute.Service
}

// RemoveIP delegates to Compute.Service.Addresses.Delete(project, region, name) function
func (caw *ComputeAPIWrapper) RemoveIP(project, region, name string) (bool, error) {
	resp, err := caw.Service.Addresses.Delete(project, region, name).Do()
	retryStatus := (resp.HTTPStatusCode != http.StatusAccepted)
	if err != nil {
		return retryStatus, err
	}
	if resp.HTTPStatusCode == http.StatusTooManyRequests {
		return retryStatus, errors.New("Quota reached")
	}

	return retryStatus, nil
}
