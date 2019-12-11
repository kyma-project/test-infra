package cluster

import (
	"context"
	"testing"
	// "errors"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster/automock"
)

var (
	testClusterName = "a-cluster"
	testClusterProj = "a-project"
	testClusterZone = "gcp-zone1-a"
)

func TestClient_Create(t *testing.T) {
	t.Run("Create() Should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		labelSet1 := make(map[string]string)
		labelSet1["test"] = "yes"

		mockAPI.On("Create", ctx, testClusterName, labelSet1, 1, true).Return(nil)

		mockClient, err := New(Option{ProjectID: testClusterProj, ZoneID: testClusterZone, ServiceAccount: "not-empty-gcp-will-validate"}, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		if err := mockClient.Create(ctx, testClusterName, labelSet1, 1, true); err != nil {
			t.Errorf("Client.Create() error = %v", err)
		}
	})
	t.Run("Create() Should throw errors because poolsize is not satisfied", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		labelSet1 := make(map[string]string)
		labelSet1["test"] = "yes"

		mockClient, err := New(Option{ProjectID: testClusterProj, ZoneID: testClusterZone, ServiceAccount: "not-empty-gcp-will-validate"}, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		if err := mockClient.Create(ctx, testClusterName, labelSet1, -1, true); err == nil {
			t.Errorf("Expecting an error but was nil %w", err)
		}

		mockAPI.AssertNumberOfCalls(t, "Create", 0)
	})
	t.Run("Create() Should throw errors because name is not satisfied", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		labelSet1 := make(map[string]string)
		labelSet1["test"] = "yes"

		mockClient, err := New(Option{ProjectID: testClusterProj, ZoneID: testClusterZone, ServiceAccount: "not-empty-gcp-will-validate"}, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		if err := mockClient.Create(ctx, "", labelSet1, -1, true); err == nil {
			t.Errorf("Client.Create() expecting an error but was nil %w", err)
		}

		mockAPI.AssertNumberOfCalls(t, "Create", 0)
	})
}

func TestClient_Delete(t *testing.T) {
	t.Run("Delete() Should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		labelSet1 := make(map[string]string)
		labelSet1["test"] = "yes"

		mockAPI.On("Delete", ctx, testClusterName).Return(nil)

		mockClient, err := New(Option{ProjectID: testClusterProj, ZoneID: testClusterZone, ServiceAccount: "not-empty-gcp-will-validate"}, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		if err := mockClient.Delete(ctx, testClusterName); err != nil {
			t.Errorf("Client.Delete() error = %v", err)
		}
	})
	t.Run("Create() Should throw errors because name is not satisfied", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		labelSet1 := make(map[string]string)
		labelSet1["test"] = "yes"

		mockClient, err := New(Option{ProjectID: testClusterProj, ZoneID: testClusterZone, ServiceAccount: "not-empty-gcp-will-validate"}, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		if err := mockClient.Delete(ctx, ""); err == nil {
			t.Errorf("Client.Delete() expecting an error but was nil %w", err)
		}

		mockAPI.AssertNumberOfCalls(t, "Delete", 0)
	})
}
