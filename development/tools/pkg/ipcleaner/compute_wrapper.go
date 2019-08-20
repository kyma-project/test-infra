package ipcleaner

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
)

const (
	ipDeletionFailed = "delete failed"
)

// ComputeAPIWrapper abstracts GCP Compute Service API
type ComputeAPIWrapper struct {
	Service *compute.Service
}

// RemoveIP delegates to Compute.Service.Addresses.Delete(project, region, name) function
func (caw *ComputeAPIWrapper) RemoveIP(ctx context.Context, project, region, name string) error {
	resp, err := caw.Service.Addresses.Delete(project, region, name).Context(ctx).Do()
	if err != nil {
		return errors.Wrap(err, "something went wrong")
	}
	if resp.HTTPStatusCode > http.StatusAccepted {
		return errors.New(ipDeletionFailed)
	}

	return nil
}
