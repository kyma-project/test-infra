package orphanremover

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/pkg/orphanremover/automock"

	"github.com/stretchr/testify/assert"
)

//filterGarbage(pool []targetPool, computeAPI ComputeAPI) []targetPool
//CheckInstance(project string, zone string, name string) bool

const testProject = "testProject"
const testZone = "europe-west3-a"

func TestToBeNamed(t *testing.T) {
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
