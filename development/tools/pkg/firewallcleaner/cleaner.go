package firewallcleaner

import (
	"context"
	"fmt"

	compute "google.golang.org/api/compute/v1"
)

const sleepFactor = 2

//go:generate mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

//ComputeAPI interface logic for Google cloud API
type ComputeAPI interface {
	DeleteHTTPProxy(project string, httpProxy string)
	DeleteURLMap(project string, urlMap string)
	DeleteBackendService(project string, backendService string)
	DeleteInstanceGroup(project string, zone string, instanceGroup string)
	DeleteHealthChecks(project string, names []string)
	DeleteForwardingRule(project string, name string, region string)
	DeleteGlobalForwardingRule(project string, name string)
	DeleteTargetPool(project string, name string, region string)
	LookupURLMaps(project string) ([]*compute.UrlMap, error)
	LookupBackendServices(project string) ([]*compute.BackendService, error)
	LookupInstanceGroup(project string, zone string) ([]string, error)
	LookupTargetPools(project string) ([]*compute.TargetPool, error)
	LookupZones(project, pattern string) ([]string, error)
	LookupHTTPProxy(project string) ([]*compute.TargetHttpProxy, error)
	LookupGlobalForwardingRule(project string) ([]*compute.ForwardingRule, error)
	CheckInstance(project string, zone string, name string) bool
}

//Cleaner Element holding the firewall cleaning logic
type Cleaner struct {
	computeAPI ComputeAPI
	githubAPI  GithubAPI
}

//NewCleaner Returns a new cleaner object
func NewCleaner(computeAPI ComputeAPI, githubAPI GithubAPI) *Cleaner {
	return &Cleaner{computeAPI, githubAPI}
}

//Run the main find&destroy function
func (c *Cleaner) Run(dryRun bool, project string) {
	ctx := context.Background()
	pulls := c.githubAPI.OpenPullRequests(ctx)

	for _, p := range pulls {
		timeClosed := p.GetClosedAt()
		fmt.Println("closed at: ", timeClosed)
	}
}
