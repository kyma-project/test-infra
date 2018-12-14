package orphanremover

import (
	"errors"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/pkg/orphanremover/automock"

	"github.com/stretchr/testify/assert"
)

//filterInstanceGroups(zones []string, computeAPI ComputeAPI, project string) []instanceGroup
//LookupInstanceGroup(project string, zone string) ([]string, error)

var (
	testProject = "testProject"
	testZones   = []string{"europe-west3-a", "europe-west3-b", "europe-west3-c"}
	testZone    = testZones[0]
)

func TestFilterGarbage(t *testing.T) {
	t.Run("filterGarbage() Should find 3 objects from 5", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		defer computeAPI.AssertExpectations(t)
		var testPool = []targetPool{
			{
				name: "usedTargetPool",
				instances: []instance{
					{
						name:   "some-vm",
						zone:   "europe-west3-a",
						exists: true,
					},
				},
				instanceCount: 1,
				healthChecks:  []string{"some-health-check"},
				region:        "europe-west3",
			},
			{
				name: "emptyTargetPool",
				instances: []instance{
					{
						name:   "some-nonexisting-vm",
						zone:   "europe-west3-a",
						exists: false,
					},
				},
				instanceCount: 1,
				healthChecks:  []string{"some-health-check"},
				region:        "europe-west3",
			},
			{
				name: "halfFullTargetPool",
				instances: []instance{
					{
						name:   "some-nonexisting-vm",
						zone:   "europe-west3-a",
						exists: false,
					},
					{
						name:   "some-vm",
						zone:   "europe-west3-a",
						exists: true,
					},
				},
				instanceCount: 2,
				healthChecks:  []string{"some-health-check"},
				region:        "europe-west3",
			},
			{
				name: "multipleHealtchchecksTargetPool",
				instances: []instance{
					{
						name:   "some-nonexisting-vm",
						zone:   "europe-west3-a",
						exists: false,
					},
				},
				instanceCount: 1,
				healthChecks:  []string{"some-health-check1", "some-health-check2"},
				region:        "europe-west3",
			},
			{
				name: "multipleEmptyTargetPool",
				instances: []instance{
					{
						name:   "some-nonexisting-vm",
						zone:   "europe-west3-a",
						exists: false,
					},
					{
						name:   "some-nonexisting-vm",
						zone:   "europe-west3-a",
						exists: false,
					},
				},
				instanceCount: 2,
				healthChecks:  []string{"some-health-check"},
				region:        "europe-west3",
			},
		}
		computeAPI.On("CheckInstance", testProject, testZone, "some-vm").Return(true)
		computeAPI.On("CheckInstance", testProject, testZone, "some-nonexisting-vm").Return(false)

		testGarbage := filterGarbage(testPool, testProject, computeAPI)
		assert.Equal(t, 3, len(testGarbage))
		for _, garbage := range testGarbage {
			assert.NotEqual(t, "halfFullTargetPool", garbage.name)
		}
		assert.Equal(t, "emptyTargetPool", testGarbage[0].name)
		assert.Equal(t, "multipleHealtchchecksTargetPool", testGarbage[1].name)
		assert.Equal(t, "multipleEmptyTargetPool", testGarbage[2].name)
	})
}

func TestFilterInstanceGroups(t *testing.T) {
	t.Run("filterInstanceGroups() Should skip unused zone", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupInstanceGroup", testProject, testZones[0]).Return([]string{"instanceGroup1", "instanceGroup2"}, nil)
		computeAPI.On("LookupInstanceGroup", testProject, testZones[1]).Return([]string{}, nil)
		computeAPI.On("LookupInstanceGroup", testProject, testZones[2]).Return([]string{"instanceGroup3"}, nil)

		filteredInstances, err := filterInstanceGroups(testZones, computeAPI, testProject)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(filteredInstances))
		for it := range filteredInstances {
			assert.NotEqual(t, "europe-west3-b", filteredInstances[it].zone)
		}
	})
	t.Run("filterInstanceGroups() Should fail when API call fails", func(t *testing.T) {
		computeAPI := &automock.ComputeAPI{}
		expectedError := errors.New("Something went wrong")
		defer computeAPI.AssertExpectations(t)

		computeAPI.On("LookupInstanceGroup", testProject, testZones[0]).Return(nil, expectedError)

		_, err := filterInstanceGroups(testZones, computeAPI, testProject)
		assert.Equal(t, expectedError, err)
	})

}
