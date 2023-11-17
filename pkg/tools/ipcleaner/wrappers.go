package ipcleaner

import (
	"context"

	"github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
)

// AddressAPIWrapper abstracts GCP Address Compute Service API
type AddressAPIWrapper struct {
	Context context.Context
	Service *compute.AddressesService
}

// ListAddresses implements AddressAPI.ListAddresses function
func (aaw *AddressAPIWrapper) ListAddresses(project, region string) ([]*compute.Address, error) {
	var addresses = []*compute.Address{}
	pageFunc := func(addressList *compute.AddressList) error {
		addresses = append(addresses, addressList.Items...)
		return nil
	}

	err := aaw.Service.List(project, region).Pages(aaw.Context, pageFunc)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

// RemoveIP delegates to Compute.Service.Addresses.Delete(project, region, name) function
func (aaw *AddressAPIWrapper) RemoveIP(project, region, name string) error {
	resp, err := aaw.Service.Delete(project, region, name).Context(aaw.Context).Do()
	if err != nil {
		return err
	}
	if resp.Error != nil {
		bytes, err := resp.Error.MarshalJSON()
		if err != nil {
			return err
		}
		return errors.New(string(bytes))
	}

	return nil
}

// RegionAPIWrapper abstracts GCP RegionsService API
type RegionAPIWrapper struct {
	Context context.Context
	Service *compute.RegionsService
}

// ListRegions implements RegionAPI.ListRegions function
func (raw *RegionAPIWrapper) ListRegions(project string) ([]string, error) {
	call := raw.Service.List(project)

	var regions []string
	f := func(page *compute.RegionList) error {
		for _, region := range page.Items {
			regions = append(regions, region.Name)
		}
		return nil
	}

	if err := call.Pages(raw.Context, f); err != nil {
		return nil, err
	}

	return regions, nil
}
