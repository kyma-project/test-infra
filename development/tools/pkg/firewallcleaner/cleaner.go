package firewallcleaner

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	compute "google.golang.org/api/compute/v1"
)

const sleepFactor = 2

//go:generate mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

//ComputeAPI interface logic for Google cloud API
type ComputeAPI interface {
	ListFirewallRules(project string) ([]*compute.Firewall, error)
	DeleteFirewallRule(project, firewall string)
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
	pulls := c.githubAPI.ClosedPullRequests(ctx)

	fwRules, err := c.computeAPI.ListFirewallRules(project)

	if err != nil {
		fmt.Println(err)
	}
	for _, p := range pulls {
		common.ShoutFirst("PR #%d: \"%s\" is %s\n", p.GetNumber(), p.GetTitle(), p.GetState())
		for _, r := range fwRules {
			prStr := fmt.Sprintf("-pr-%d", p.GetNumber())
			if strings.Contains(r.Name, prStr) {
				// c.computeAPI.DeleteFirewallRule(project, r.Name)
				common.Shout("If I were serious, I'd delete the rule for the above PR here because of the name. Rule name: %s", r.Name)
			}

			//
			// TODO: Is this necessary? Do target tags get updated, if there's one rule for multiple targets and one target gets deleted?
			// If target tags don't get updated workflow should be, check if target tag is the last one remaining, if not, remove this rules target tag from all rules found
			// If target tags shouldn't be considered, delete this codeblock below
			for _, t := range r.TargetTags {
				if strings.Contains(t, prStr) {
					// c.computeAPI.DeleteFirewallRule(project, r.Name)
					common.Shout("If I were serious, I'd delete the rule for the above PR here, because of the target tag. Rule name: %s, TargetTag: %s", r.Name, t)
				}
			}
		}
	}
}
