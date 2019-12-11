package cluster

import (
	"context"
	"reflect"
	"testing"

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

func TestNew(t *testing.T) {
	t.Run("New() should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)
		opts := Option{ProjectID: "string", ZoneID: "string", ServiceAccount: "string"}

		c, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("New() error = %v, should've not thrown an error", err)
		}
		want := &Client{Option: opts, api: mockAPI}
		if !reflect.DeepEqual(c, want) {
			t.Errorf("New() %v, want = %v", c, want)
		}
		mockAPI.AssertNumberOfCalls(t, "Create", 0)
		mockAPI.AssertNumberOfCalls(t, "Delete", 0)
	})
	t.Run("New() should throw errors, because ProjectID is not satisfied", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)
		opts := Option{ProjectID: "", ZoneID: "string", ServiceAccount: "string"}

		c, err := New(opts, mockAPI)
		if err == nil {
			t.Errorf("New() expected an error")
		}
		if c != nil {
			t.Errorf("New() %v, want = %v", c, nil)
		}
		mockAPI.AssertNumberOfCalls(t, "Create", 0)
		mockAPI.AssertNumberOfCalls(t, "Delete", 0)
	})
	t.Run("New() should throw errors, because ZoneID is not satisfied", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)
		opts := Option{ProjectID: "string", ZoneID: "", ServiceAccount: "string"}

		c, err := New(opts, mockAPI)
		if err == nil {
			t.Errorf("New() expected an error")
		}
		if c != nil {
			t.Errorf("New() %v, want = %v", c, nil)
		}
		mockAPI.AssertNumberOfCalls(t, "Create", 0)
		mockAPI.AssertNumberOfCalls(t, "Delete", 0)
	})
	t.Run("New() should throw errors, because ServiceAccount is not satisfied", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)
		opts := Option{ProjectID: "string", ZoneID: "string", ServiceAccount: ""}

		c, err := New(opts, mockAPI)
		if err == nil {
			t.Errorf("New() expected an error")
		}
		if c != nil {
			t.Errorf("New() %v, want = %v", c, nil)
		}
		mockAPI.AssertNumberOfCalls(t, "Create", 0)
		mockAPI.AssertNumberOfCalls(t, "Delete", 0)
	})
}
