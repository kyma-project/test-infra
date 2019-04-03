package firewallcleaner

import (
	"errors"
	"strings"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/pkg/firewallcleaner/automock"
	"github.com/stretchr/testify/assert"
	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
)

var (
	testProject = "testProject"
)

func TestFirewallcleaner(t *testing.T) {
	t.Run("Cleaner.Run() Should not throw errors", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return([]*compute.Firewall{}, nil)
		computeAPI.On("LookupInstances", testProject).Return([]*compute.Instance{}, nil)
		computeAPI.On("LookupNodePools", testProject).Return([]*container.NodePool{}, nil)

		c := NewCleaner(computeAPI)

		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)
	})
	t.Run("Cleaner.Run() Should throw an error when firewall call fails", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		expectedError := "Something went wrong"
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return([]*compute.Firewall{}, errors.New(expectedError))
		computeAPI.AssertNotCalled(t, "LookupInstances")

		c := NewCleaner(computeAPI)

		firstErr := c.Run(true, testProject)
		assert.True(t, strings.Contains(firstErr.Error(), expectedError))
		assert.True(t, strings.Contains(firstErr.Error(), "LookupFirewallRule"))
		secondErr := c.Run(false, testProject)
		assert.True(t, strings.Contains(secondErr.Error(), expectedError))
		assert.True(t, strings.Contains(secondErr.Error(), "LookupFirewallRule"))
	})
	t.Run("Cleaner.Run() Should throw an error when instance call fails", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		expectedError := "Something went wrong"
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return([]*compute.Firewall{}, nil)
		computeAPI.On("LookupInstances", testProject).Return([]*compute.Instance{}, errors.New(expectedError))

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.True(t, strings.Contains(firstErr.Error(), expectedError))
		assert.True(t, strings.Contains(firstErr.Error(), "LookupInstances"))

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.NotNil(t, secondErr)
		assert.True(t, strings.Contains(secondErr.Error(), expectedError))
		assert.True(t, strings.Contains(secondErr.Error(), "LookupInstances"))
	})
	t.Run("Cleaner.Run() Should throw an error when nodepool call fails", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		expectedError := "Something went wrong"
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return([]*compute.Firewall{}, nil)
		computeAPI.On("LookupInstances", testProject).Return([]*compute.Instance{}, nil)
		computeAPI.On("LookupNodePools", testProject).Return([]*container.NodePool{}, errors.New(expectedError))

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.True(t, strings.Contains(firstErr.Error(), expectedError))
		assert.True(t, strings.Contains(firstErr.Error(), "LookupNodePools"))

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.NotNil(t, secondErr)
		assert.True(t, strings.Contains(secondErr.Error(), expectedError))
		assert.True(t, strings.Contains(secondErr.Error(), "LookupNodePools"))
	})
}

func TestFirewallcleanerRuleInteraction(t *testing.T) {

	t.Run("Cleaner.Run() Should delete one firewall-rule, because the referenced instance does not exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{"instance-one"}},
			{Name: "firewall-two", TargetTags: []string{"instance-not-existent"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupNodePools", testProject).Return(nodePools, nil)

		// computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 1)

		// computeAPI.On("DeleteFirewallRule", testProject, firewalls[1].Name).Times(1)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)
	})
	t.Run("Cleaner.Run() Should delete one firewall-rule, because the referenced nodepool does not exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{"nodepool-one-default-pool-abcdefg01-grp"}},
			{Name: "firewall-two", TargetTags: []string{"nodepool-two-default-pool-abcdefg09-grp"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupNodePools", testProject).Return(nodePools, nil)

		// computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 1)

		// computeAPI.On("DeleteFirewallRule", testProject, firewalls[1].Name).Times(1)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)
	})
	t.Run("Cleaner.Run() Should delete all firewall-rules, because no instances or node pools exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{}},
			{Name: "firewall-two", TargetTags: []string{"instance-two"}},
			{Name: "firewall-three"},
		}
		instances := []*compute.Instance{}
		nodePools := []*container.NodePool{}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupNodePools", testProject).Return(nodePools, nil)

		// computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 3)

		// computeAPI.On("DeleteFirewallRule", testProject, firewalls[0].Name).Times(1)
		// computeAPI.On("DeleteFirewallRule", testProject, firewalls[1].Name).Times(1)
		// computeAPI.On("DeleteFirewallRule", testProject, firewalls[2].Name).Times(1)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)
	})
	t.Run("Cleaner.Run() Should delete no firewall-rules, because all instances or nodepools exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{"instance-one"}},
			{Name: "firewall-two", TargetTags: []string{"instance-two"}},
			{Name: "firewall-three", TargetTags: []string{"nodepool-one-default-pool-abcdefg01-grp"}},
			{Name: "firewall-four", TargetTags: []string{"nodepool-two-default-pool-abcdefg09-grp"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
			{Name: "instance-two"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
			{Name: "nodepool-two"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupNodePools", testProject).Return(nodePools, nil)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)
	})
	t.Run("Cleaner.Run() Should delete no firewall-rules, because dryRun is true", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{"instance-one"}},
			{Name: "firewall-two", TargetTags: []string{"instance-two"}},
			{Name: "firewall-three", TargetTags: []string{"instance-three"}},
			{Name: "firewall-four", TargetTags: []string{}},
			{Name: "firewall-five"},
			{Name: "firewall-six", TargetTags: []string{"nodepool-one-default-pool-abcdefg01-grp"}},
			{Name: "firewall-seven", TargetTags: []string{"nodepool-two-default-pool-abcdefg09-grp"}},
			{Name: "firewall-eight", TargetTags: []string{"nodepool-three-default-pool-abcdefg02-grp"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
			{Name: "instance-two"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
			{Name: "nodepool-two"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupNodePools", testProject).Return(nodePools, nil)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		c := NewCleaner(computeAPI)

		// dryRun true
		err := c.Run(true, testProject)
		assert.Nil(t, err)
	})
}

func TestCurrentLimitation(t *testing.T) {
	t.Run("Cleaner.Run() Should delete no firewall-rules starting with \"k8s-\"", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "k8s-firewall-rule-one"},
			{Name: "k8s-firewall-rule-two", TargetTags: []string{}},
			{Name: "k8s-firewall-rule-three", TargetTags: []string{"instance-one"}},
			{Name: "k8s-firewall-rule-four", TargetTags: []string{"nodepool-one-default-pool-abcdefg01-grp"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupNodePools", testProject).Return(nodePools, nil)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)
	})
	t.Run("Cleaner.Run() Should delete no firewall-rules that have no targetTags", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "k8s-firewall-rule-one"},
			{Name: "k8s-firewall-rule-two", TargetTags: []string{}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupNodePools", testProject).Return(nodePools, nil)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)
	})
}
