package dnsclient

import (
	"context"
	"time"
)

const (
	// DefaultPropagationWait is the default propagation waiting time.
	DefaultPropagationWait = 60 * time.Second

	// DefaultCheckDelay is the default check delay
	DefaultCheckDelay = 100 * time.Millisecond

	// DefaultProvisionDelay is the default after provision wait delay.
	DefaultProvisionDelay = 10 * time.Second

	// DefaultDNSProject is a default GCP project which host DNS Managed Zone.
	DefaultDNSProject = "sap-kyma-prow-workloads"

	// DefaultZoneDNSName is a default zone DNS name which contains DNS records.
	DefaultZoneName = "build-kyma-workloads"
)

type DNSClient struct {
	service DNSAPI
}

type DNSAPI interface {
	GetManagedZone(ctx context.Context, project string, managedZone string) (*dns.ManagedZone, error)
	ChangeRecord(ctx context.Context, project string, managedZone string, change *dns.Change) (*dns.Change, error)
}

type RecordOpts struct {
	project    string
	zoneName   string
	name       string
	data       string
	recordType string
	ttl        int64
}

type DNSChange struct {
	change *dns.Change
	rrs    *dns.ResourceRecordSet
	opts   RecordOpts
}

// New return new DNS client and error object. Error is not used at present. Added it for future use and to support common error handling.
// Call NewService to get API implementation.
func New(service DNSAPI) (*DNSClient, error) {
	return &DNSClient{
		service: service,
	}, nil
}

func NewRecordOpts(project, zoneName, name, data, recordType string, ttl int64) RecordOpts {
	if project == "" {
		project = DefaultDNSProject
	}
	if zoneName == "" {
		zoneName = DefaultZoneName
	}
	return RecordOpts{
		project:    project,
		zoneName:   zoneName,
		name:       name,
		data:       data,
		recordType: recordType,
		ttl:        ttl,
	}
}

func (client *DNSClient) NewDNSChange(record RecordOpts) *DNSChange {
	return &DNSChange{
		change: &dns.Change{
			IsServing: true,
		},
		rrs: &dns.ResourceRecordSet{
			Name:    record.name,
			Rrdatas: []string{record.data},
			Type:    record.recordType,
			Ttl:     record.ttl,
		},
		opts: record,
	}
}

func (dnschange *DNSChange) AddRecord() *DNSChange {
	dnschange.change.Additions = append(dnschange.change.Additions, dnschange.rrs)
	return dnschange
}

func (dnschange *DNSChange) DeleteRecord() *DNSChange {
	dnschange.change.Deletions = append(dnschange.change.Deletions, dnschange.rrs)
	return dnschange
}

func (client *DNSClient) DoChange(ctx context.Context, dnschange *DNSChange) (*dns.Change, error) {
	change, changeErr := client.service.ChangeRecord(ctx, dnschange.opts.project, dnschange.opts.zoneName, dnschange.change)
	if changeErr != nil {
		return nil, changeErr
	}
	return change, nil
}

func (client *DNSClient) LookupDNSRecord(ctx context.Context, record RecordOpts) (*dns.ResourceRecordSet, error) {
	zone, err := client.FindManagedZone(ctx, record)
	if err != nil {
		return nil, err
	}
	record.zoneName = zone.Name
	return nil, nil
}

// FindZone will search for managed zone closest to RecordOpts.
func (client *DNSClient) FindManagedZone(ctx context.Context, record RecordOpts) (*dns.ManagedZone, error) {
	var (
		zone    *dns.ManagedZone
		zoneErr error
	)
	if record.zoneName != "" {
		zone, zoneErr = client.service.GetManagedZone(ctx, record.project, record.zoneName)
		if zoneErr != nil {
			return nil, zoneErr
		}
	}
	return zone, nil
}
