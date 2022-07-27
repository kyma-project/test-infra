package diskscollector

import (
	"context"
	"errors"

	compute "google.golang.org/api/compute/v1"
)

// DiskAPIWrapper abstracts GCP DisksService API
type DiskAPIWrapper struct {
	Context context.Context
	Service *compute.DisksService
}

// ListDisks implements DiskAPI.ListDisks function
func (daw *DiskAPIWrapper) ListDisks(project, zone string) ([]*compute.Disk, error) {

	var disks = []*compute.Disk{}
	pageFunc := func(disksList *compute.DiskList) error {
		disks = append(disks, disksList.Items...)
		return nil
	}

	err := daw.Service.List(project, zone).Pages(daw.Context, pageFunc)
	if err != nil {
		return nil, err
	}
	return disks, nil
}

// RemoveDisk implements DiskAPI.RemoveDisk function
func (daw *DiskAPIWrapper) RemoveDisk(project, zone, name string) error {

	operation, err := daw.Service.Delete(project, zone, name).Context(daw.Context).Do()

	if err != nil {
		return err
	}

	if operation.Error != nil {
		bytes, err := operation.Error.MarshalJSON()
		if err != nil {
			return err
		}
		return errors.New(string(bytes))
	}

	return nil
}

// ZoneAPIWrapper abstracts GCP ZonesService API
type ZoneAPIWrapper struct {
	Context context.Context
	Service *compute.ZonesService
}

// ListZones implements ZoneAPI.ListZones function
func (zaw *ZoneAPIWrapper) ListZones(project string) ([]string, error) {
	call := zaw.Service.List(project)

	var zones []string
	f := func(page *compute.ZoneList) error {
		for _, zone := range page.Items {
			zones = append(zones, zone.Name)
		}
		return nil
	}

	if err := call.Pages(zaw.Context, f); err != nil {
		return nil, err
	}

	return zones, nil
}
