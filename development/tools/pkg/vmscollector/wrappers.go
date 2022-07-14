package vmscollector

import (
	"context"

	compute "google.golang.org/api/compute/v1"
)

// InstancesAPIWrapper abstracts GCP InstancesService API
type InstancesAPIWrapper struct {
	Context context.Context
	Service *compute.InstancesService
}

// ListInstances delegates to InstancesService.ListInstances(project) function
func (iaw *InstancesAPIWrapper) ListInstances(project string) ([]*compute.Instance, error) {

	var instances = []*compute.Instance{}

	pageFunc := func(instancesList *compute.InstanceAggregatedList) error {
		for _, instancesInZone := range instancesList.Items {
			instances = append(instances, instancesInZone.Instances...)
		}
		return nil
	}

	err := iaw.Service.AggregatedList(project).Pages(iaw.Context, pageFunc)

	if err != nil {
		return nil, err
	}

	return instances, nil
}

// RemoveInstance delegates to InstancesService.Delete(project, zone, name) function
func (iaw *InstancesAPIWrapper) RemoveInstance(project, zone, name string) error {

	_, err := iaw.Service.Delete(project, zone, name).Context(iaw.Context).Do()

	if err != nil {
		return err
	}
	return nil
}
