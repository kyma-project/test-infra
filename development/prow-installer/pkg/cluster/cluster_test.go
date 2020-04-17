package cluster

import (
	"context"
	"testing"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster/automock"

	"github.com/stretchr/testify/assert"
)

var (
	opts = Option{
		Prefix:         "test-prefix",
		ProjectID:      "a-project",
		ServiceAccount: "not-empty-gcp-will-validate",
	}
)

func TestClient_Create(t *testing.T) {
	t.Run("Create() Should not throw errors", func(t *testing.T) {
		ctx := context.Background()
		api := automock.API{}
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
		client, err := New(opts, &api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Create(ctx, testClusterConfig)
		assert.NoErrorf(t, err, "Create returned no errors for valid config")
	})
	t.Run("Create() Should throw errors because initialSize is not satisfied", func(t *testing.T) {
		ctx := context.Background()
		api := automock.API{}
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
		client, err := New(opts, &api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Create(ctx, testClusterConfig)
		assert.EqualErrorf(t, err, "error creating node pool configuration size must be at least 1", "size is not satisfied")
	})
	t.Run("Create() Should throw errors because name is not satisfied", func(t *testing.T) {
		ctx := context.Background()
		api := automock.API{}
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
		client, err := New(opts, &api)
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
		api := &automock.API{}

		client, err := New(opts, api)
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
		api := &automock.API{}

		client, err := New(opts, api)
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
		api := &automock.API{}

		client, err := New(opts, api)
		if err != nil {
			t.Errorf("error ocured during client creation")
		}
		err = client.Delete(ctx, testClusterName, testZoneId)
		assert.EqualErrorf(t, err, "zoneId cannot be empty", "name is not satisfied in Delete()")
	})
}

func TestNew(t *testing.T) {
	t.Run("New() should not throw errors", func(t *testing.T) {
		api := &automock.API{}
		_, err := New(opts, api)
		assert.NoErrorf(t, err, "no errors during client creation")
	})
	t.Run("New() should throw errors, because ProjectID is not satisfied", func(t *testing.T) {
		api := &automock.API{}
		testOpts := &Option{
			Prefix:         "test-prefix",
			ProjectID:      "",
			ServiceAccount: "gke-test-serviceaccount",
		}
		_, err := New(*testOpts, api)
		assert.EqualErrorf(t, err, "ProjectID is required to initialize a client", "ProjectID is not satisfied in New()")
	})
	t.Run("New() should throw errors, because ServiceAccount is not satisfied", func(t *testing.T) {
		api := &automock.API{}
		testOpts := &Option{
			Prefix:         "test-prefix",
			ProjectID:      "gcp-test-project",
			ServiceAccount: "",
		}
		_, err := New(*testOpts, api)
		assert.EqualErrorf(t, err, "ServiceAccount is required to initialize a client", "ServiceAccount is not satisfied in New()")
	})

	t.Run("New() should throw errors, because api is not initialized", func(t *testing.T) {
		testOpts := &Option{
			Prefix:         "test-prefix",
			ProjectID:      "gcp-test-project",
			ServiceAccount: "gke-test-serviceaccount",
		}
		_, err := New(*testOpts, nil)
		assert.EqualErrorf(t, err, "api is required to initialize a client", "api is not satisfied in New()")
	})
}
