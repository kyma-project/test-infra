package dnsclient

import (
	"context"
	"google.golang.org/api/dns/v2beta1"
)

type DNSAPIWrapper struct {
	service *dns.Service
}

func NewService(ctx context.Context) (*DNSAPIWrapper, error) {
	service, err := dns.NewService(ctx)
	if err != nil {
		return nil, err
	} else {
		return &DNSAPIWrapper{
			service: service,
		}, err
	}
}

func (api *DNSAPIWrapper) GetManagedZone(ctx context.Context, project string, managedZone string) (*dns.ManagedZone, error) {
	return api.service.ManagedZones.Get(project, managedZone).Context(ctx).Do()
}

func (api *DNSAPIWrapper) ChangeRecord(ctx context.Context, project string, managedZone string, change *dns.Change) (*dns.Change, error) {
	return api.service.Changes.Create(project, managedZone, change).Context(ctx).Do()
}
