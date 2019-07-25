package ipcleaner

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
)

// ComputeAPIWrapper abstracts GCP Compute Service API
type ComputeAPIWrapper struct {
	Context context.Context
	Service *compute.Service
}

// RemoveIP delegates to Compute.Service.Addresses.Delete(project, region, name) function
func (caw *ComputeAPIWrapper) RemoveIP(project, region, name string) error {
	resp, err := caw.Service.Addresses.Delete(project, region, name).Context(caw.Context).Do()
	if err != nil {
		if resp != nil {
			switch resp.HTTPStatusCode {
			case http.StatusNotFound:
				return errors.Wrap(err, string(http.StatusNotFound))
			}
		}
		return errors.Wrap(err, "something went wrong")
	}

	return nil
}
