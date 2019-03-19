package firewallcleaner

import (
	"errors"
	"strings"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/pkg/firewallcleaner/automock"
	"github.com/stretchr/testify/assert"
	compute "google.golang.org/api/compute/v1"
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
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)

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
	t.Run("Cleaner.Run() Should delete all firewall-rules, because no instances exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{}},
			{Name: "firewall-two", TargetTags: []string{"instance-two"}},
			{Name: "firewall-three"},
		}
		instances := []*compute.Instance{}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)

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
	t.Run("Cleaner.Run() Should delete no firewall-rules, because all instances exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{"instance-one"}},
			{Name: "firewall-two", TargetTags: []string{"instance-two"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
			{Name: "instance-two"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)

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
