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
		computeAPI.On("LookupClusters", testProject).Return([]*container.Cluster{}, nil)
		computeAPI.On("LookupNodePools", []*container.Cluster{}).Return([]*container.NodePool{}, nil)

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
	t.Run("Cleaner.Run() Should throw an error when cluster call fails", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		expectedError := "Something went wrong"
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return([]*compute.Firewall{}, nil)
		computeAPI.On("LookupInstances", testProject).Return([]*compute.Instance{}, nil)
		computeAPI.On("LookupClusters", testProject).Return([]*container.Cluster{}, errors.New(expectedError))

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.True(t, strings.Contains(firstErr.Error(), expectedError))
		assert.True(t, strings.Contains(firstErr.Error(), "LookupClusters"))

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.NotNil(t, secondErr)
		assert.True(t, strings.Contains(secondErr.Error(), expectedError))
		assert.True(t, strings.Contains(secondErr.Error(), "LookupClusters"))
	})
	t.Run("Cleaner.Run() Should throw an error when nodepool call fails", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		expectedError := "Something went wrong"
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return([]*compute.Firewall{}, nil)
		computeAPI.On("LookupInstances", testProject).Return([]*compute.Instance{}, nil)
		computeAPI.On("LookupClusters", testProject).Return([]*container.Cluster{}, nil)
		computeAPI.On("LookupNodePools", []*container.Cluster{}).Return([]*container.NodePool{}, errors.New(expectedError))

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
		clusters := []*container.Cluster{
			{Name: "cluster-name"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupClusters", testProject).Return(clusters, nil)
		computeAPI.On("LookupNodePools", clusters).Return(nodePools, nil)
		computeAPI.On("DeleteFirewallRule", testProject, "firewall-two")

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 1)
	})
	t.Run("Cleaner.Run() Should delete one firewall-rule, because the referenced nodepool does not exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one-nodepool-one", TargetTags: []string{"nodepool-one-default-pool-abcdefg01-grp"}},
			{Name: "firewall-two-nodepool-two", TargetTags: []string{"nodepool-two-default-pool-abcdefg09-grp"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
		}
		clusters := []*container.Cluster{
			{Name: "cluster-name"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one", InitialNodeCount: 1, InstanceGroupUrls: []string{"https://www.googleapis.com/compute/v1/projects/sap-kyma-prow/zones/europe-west3-a/instanceGroupManagers/nodepool-one-default-pool-abcdefg01-grp"}},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupClusters", testProject).Return(clusters, nil)
		computeAPI.On("LookupNodePools", clusters).Return(nodePools, nil)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		computeAPI.On("DeleteFirewallRule", testProject, "firewall-two-nodepool-two")

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 1)
	})
	t.Run("Cleaner.Run() Should delete one firewall-rule, because the referenced cluster does not exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{"cluster-name-commit-something-something-node"}},
			{Name: "firewall-two", TargetTags: []string{"cluster-name-pr-3451-something-something-node"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
		}
		clusters := []*container.Cluster{
			{Name: "cluster-name-pr-3451-something"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupClusters", testProject).Return(clusters, nil)
		computeAPI.On("LookupNodePools", clusters).Return(nodePools, nil)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		computeAPI.On("DeleteFirewallRule", testProject, "firewall-one")

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 1)
	})
	t.Run("Cleaner.Run() Should delete one firewall-rule, because no instances or node pools or clusters exist for that rule and the other rules have no target tags so they're ignored", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{}},
			{Name: "firewall-two", TargetTags: []string{"instance-two"}},
			{Name: "firewall-three"},
		}
		instances := []*compute.Instance{}
		clusters := []*container.Cluster{}
		nodePools := []*container.NodePool{}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupClusters", testProject).Return(clusters, nil)
		computeAPI.On("LookupNodePools", clusters).Return(nodePools, nil)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		computeAPI.On("DeleteFirewallRule", testProject, firewalls[1].Name).Return()

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 1)
	})
	t.Run("Cleaner.Run() Should delete no firewall-rules, because all instances or nodepools or clusters exist", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "firewall-one", TargetTags: []string{"instance-one"}},
			{Name: "firewall-two", TargetTags: []string{"instance-two"}},
			{Name: "firewall-three-nodepool-one", TargetTags: []string{"nodepool-one-default-pool-abcdefg01-grp"}},
			{Name: "firewall-four-nodepool-two", TargetTags: []string{"nodepool-two-default-pool-abcdefg09-grp"}},
			{Name: "firewall-five", TargetTags: []string{"cluster-name-commit-something-something-node"}},
			{Name: "firewall-six", TargetTags: []string{"cluster-name-pr-3451-something-something-node"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
			{Name: "instance-two"},
		}
		clusters := []*container.Cluster{
			{Name: "cluster-name-commit-something"},
			{Name: "cluster-name-pr-3451-something"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one", InitialNodeCount: 1, InstanceGroupUrls: []string{"https://www.googleapis.com/compute/v1/projects/sap-kyma-prow/zones/europe-west3-a/instanceGroupManagers/nodepool-one-default-pool-abcdefg01-grp"}},
			{Name: "nodepool-two", InitialNodeCount: 1, InstanceGroupUrls: []string{"https://www.googleapis.com/compute/v1/projects/sap-kyma-prow/zones/europe-west3-a/instanceGroupManagers/nodepool-two-default-pool-abcdefg09-grp"}},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupClusters", testProject).Return(clusters, nil)
		computeAPI.On("LookupNodePools", clusters).Return(nodePools, nil)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)
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
			{Name: "firewall-nine", TargetTags: []string{"cluster-name-commit-something-something-node"}},
			{Name: "firewall-ten", TargetTags: []string{"cluster-name-pr-3451-something-something-node"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-one"},
			{Name: "instance-two"},
		}
		clusters := []*container.Cluster{
			{Name: "cluster-name-commit-something"},
			{Name: "cluster-name-pr-3451-something"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
			{Name: "nodepool-two"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupClusters", testProject).Return(clusters, nil)
		computeAPI.On("LookupNodePools", clusters).Return(nodePools, nil)

		c := NewCleaner(computeAPI)

		// dryRun true
		err := c.Run(true, testProject)
		assert.Nil(t, err)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)
	})
}

func TestK8sNames(t *testing.T) {
	t.Run("Cleaner.Run() Should not ignore firewall-rules starting with \"k8s-\"", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		firewalls := []*compute.Firewall{
			{Name: "k8s-firewall-rule-three", TargetTags: []string{"instance-one"}},
			{Name: "k8s-firewall-rule-four", TargetTags: []string{"nodepool-one-default-pool-abcdefg01-grp"}},
			{Name: "k8s-firewall-rule-five", TargetTags: []string{"cluster-name-commit-something-something-node"}},
			{Name: "k8s-firewall-rule-six", TargetTags: []string{"cluster-name-pr-3451-something-something-node"}},
		}
		instances := []*compute.Instance{
			{Name: "instance-two"},
		}
		clusters := []*container.Cluster{
			{Name: "cluster-one"},
			{Name: "cluster-two"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-two"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupClusters", testProject).Return(clusters, nil)
		computeAPI.On("LookupNodePools", clusters).Return(nodePools, nil)
		computeAPI.On("DeleteFirewallRule", testProject, "k8s-firewall-rule-three").Return()
		computeAPI.On("DeleteFirewallRule", testProject, "k8s-firewall-rule-four").Return()
		computeAPI.On("DeleteFirewallRule", testProject, "k8s-firewall-rule-five").Return()
		computeAPI.On("DeleteFirewallRule", testProject, "k8s-firewall-rule-six").Return()

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 4)
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
		clusters := []*container.Cluster{
			{Name: "cluster-one"},
		}
		nodePools := []*container.NodePool{
			{Name: "nodepool-one"},
		}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupFirewallRule", testProject).Return(firewalls, nil)
		computeAPI.On("LookupInstances", testProject).Return(instances, nil)
		computeAPI.On("LookupClusters", testProject).Return(clusters, nil)
		computeAPI.On("LookupNodePools", clusters).Return(nodePools, nil)

		c := NewCleaner(computeAPI)

		// dryRun true
		firstErr := c.Run(true, testProject)
		assert.Nil(t, firstErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)

		// dryRun false
		secondErr := c.Run(false, testProject)
		assert.Nil(t, secondErr)

		computeAPI.AssertNumberOfCalls(t, "DeleteFirewallRule", 0)
	})
}
