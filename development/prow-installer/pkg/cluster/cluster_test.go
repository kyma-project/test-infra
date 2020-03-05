package cluster

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	prefix    = "test-prefix"
	projectID = "a-project"
)

func TestClient_Create(t *testing.T) {
	t.Run("Create() Should not throw errors", func(t *testing.T) {
		ctx := context.Background()
		api := MockAPI{}
		testClusterConfig := Cluster{
			Name:        "test-cluster-name",
			Location:    "gcp-zone1-a",
			Description: "Test description",
			Labels:      map[string]string{"test": "yes"},
			Pools: []Pool{
				{
					Name: "test-pool",
					Size: 2,
					NodeConfig: NodeConfig{
						MachineType: "n1-standard-1",
						DiskType:    "pd-standard",
						DiskSizeGb:  100,
					},
					Autoscaling: Autoscaling{
						Enabled:      true,
						MinNodeCount: 1,
						MaxNodeCount: 5,
					},
				},
			},
		}
		client, err := New(projectID, prefix, &api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Create(ctx, testClusterConfig)
		assert.NoErrorf(t, err, "Create returned no errors for valid config")
	})
	t.Run("Create() Should throw errors because initialSize is not satisfied", func(t *testing.T) {
		ctx := context.Background()
		api := MockAPI{}
		testClusterConfig := Cluster{
			Name:        "test-cluster-name",
			Location:    "gcp-zone1-a",
			Description: "Test description",
			Labels:      map[string]string{"test": "yes"},
			Pools: []Pool{
				{
					Name: "test-pool",
					Size: 0,
					NodeConfig: NodeConfig{
						MachineType: "n1-standard-1",
						DiskType:    "pd-standard",
						DiskSizeGb:  100,
					},
					Autoscaling: Autoscaling{
						Enabled:      true,
						MinNodeCount: 1,
						MaxNodeCount: 5,
					},
				},
			},
		}
		client, err := New(projectID, prefix, &api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Create(ctx, testClusterConfig)
		assert.EqualErrorf(t, err, "error creating node pool configuration size must be at least 1", "size is not satisfied")
	})
	t.Run("Create() Should throw errors because name is not satisfied", func(t *testing.T) {
		ctx := context.Background()
		api := MockAPI{}
		testClusterConfig := Cluster{
			Name:        "",
			Location:    "gcp-zone1-a",
			Description: "Test description",
			Labels:      map[string]string{"test": "yes"},
			Pools: []Pool{
				{
					Name: "test-pool",
					Size: 0,
					NodeConfig: NodeConfig{
						MachineType: "n1-standard-1",
						DiskType:    "pd-standard",
						DiskSizeGb:  100,
					},
					Autoscaling: Autoscaling{
						Enabled:      true,
						MinNodeCount: 1,
						MaxNodeCount: 5,
					},
				},
			},
		}
		client, err := New(projectID, prefix, &api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Create(ctx, testClusterConfig)
		assert.EqualErrorf(t, err, "name cannot be empty", "name is not satisfied in Create()")
	})
}

func TestClient_Delete(t *testing.T) {
	t.Run("Delete() Should not throw errors", func(t *testing.T) {
		testClusterName := "test-cluster-name"
		testZoneId := "gcp-zone1-a"
		ctx := context.Background()
		api := MockAPI{}

		client, err := New(projectID, prefix, &api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Delete(ctx, testClusterName, testZoneId)
		assert.NoErrorf(t, err, "no errors on Delete")
	})
	t.Run("Delete() Should throw errors because name is not satisfied", func(t *testing.T) {
		testClusterName := ""
		testZoneId := "gcp-zone1-a"
		ctx := context.Background()
		api := MockAPI{}

		client, err := New(projectID, prefix, &api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Delete(ctx, testClusterName, testZoneId)
		assert.EqualErrorf(t, err, "name cannot be empty", "name is not satisfied in Delete()")
	})
	t.Run("Delete() Should throw errors because zoneId is not satisfied", func(t *testing.T) {
		testClusterName := "test-cluster-name"
		testZoneId := ""
		ctx := context.Background()
		api := MockAPI{}

		client, err := New(projectID, prefix, &api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Delete(ctx, testClusterName, testZoneId)
		assert.EqualErrorf(t, err, "zoneId cannot be empty", "name is not satisfied in Delete()")
	})
}

func TestNew(t *testing.T) {
	t.Run("New() should not throw errors", func(t *testing.T) {
		api := MockAPI{}
		_, err := New(projectID, prefix, &api)
		assert.NoErrorf(t, err, "no errors during client creation")
	})
	t.Run("New() should throw errors, because projectID is not satisfied", func(t *testing.T) {
		api := MockAPI{}
		fakeProjectId := ""
		_, err := New(fakeProjectId, prefix, &api)
		assert.EqualErrorf(t, err, "projectID is required to initialize a client", "projectID is not satisfied in New()")
	})

	t.Run("New() should throw errors, because api is not initialized", func(t *testing.T) {
		_, err := New(projectID, prefix, nil)
		assert.EqualErrorf(t, err, "api is required to initialize a client", "api is not satisfied in New()")
	})
}
