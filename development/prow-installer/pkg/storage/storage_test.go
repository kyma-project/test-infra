package storage

import (
	"context"
	"testing"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage/automock"
	"github.com/stretchr/testify/assert"
)

var (
	testGCSProj          = "gcs-test-project"
	testGCSLocation      = "gcs-test-location"
	testGCSPrefix        = "gcs-test-prefix"
	testGCSBucket        = "gcs-test-bucket"
	testGCSStorageObject = "gcs-test-storage-object"

	testBucketContent = "gcs-test-bucket-content"
)

func TestClient_Read(t *testing.T) {
	t.Run("Read() Should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockAPI.On("Read", ctx, testGCSBucket, testGCSStorageObject).Return([]byte(testBucketContent), nil)

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		data, err := mockClient.Read(ctx, testGCSBucket, testGCSStorageObject)
		if err != nil {
			t.Errorf("Client.Read() error = %v", err)
		}
		assert.Equal(t, string(data), testBucketContent)
	})
	t.Run("Read() Should throw error when bucket is not passed", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		data, err := mockClient.Read(ctx, "", testGCSStorageObject)
		if err == nil {
			t.Errorf("Client.Read() should have thrown an error")
		}
		assert.Nil(t, data)
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
	})
	t.Run("Read() Should throw error when storage object is not passed", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		data, err := mockClient.Read(ctx, testGCSBucket, "")
		if err == nil {
			t.Errorf("Client.Read() should have thrown an error")
		}
		assert.Nil(t, data)
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
	})
}

func TestClient_Write(t *testing.T) {
	t.Run("Write() Should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockAPI.On("Write", ctx, []byte(testBucketContent), testGCSBucket, testGCSStorageObject).Return(nil)

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		err = mockClient.Write(ctx, []byte(testBucketContent), testGCSBucket, testGCSStorageObject)
		if err != nil {
			t.Errorf("Client.Write() error = %v", err)
		}
	})
	t.Run("Write() Should throw error when bucket is not passed", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		err = mockClient.Write(ctx, []byte(testBucketContent), "", testGCSStorageObject)
		if err == nil {
			t.Errorf("Client.Write() should have thrown an error")
		}
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
	})
	t.Run("Write() Should throw error when storage object is not passed", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		err = mockClient.Write(ctx, []byte(testBucketContent), testGCSBucket, "")
		if err == nil {
			t.Errorf("Client.Write() should have thrown an error")
		}
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
	})
	t.Run("Write() Should throw error when len(0) byte is passed", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		err = mockClient.Write(ctx, []byte(""), testGCSBucket, testGCSStorageObject)
		if err == nil {
			t.Errorf("Client.Write() should have thrown an error")
		}
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
	})
}

func TestNew(t *testing.T) {
	t.Run("New() should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient == nil {
			t.Errorf("New() expected client to not be nil")
			t.Fail()
		}
		if err != nil {
			t.Errorf("New() error should be nil %v", err)
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
		mockAPI.AssertNumberOfCalls(t, "CreateBucket", 0)
		mockAPI.AssertNumberOfCalls(t, "DeleteBucket", 0)
	})
	t.Run("New() Should throw error when project id is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
		mockAPI.AssertNumberOfCalls(t, "CreateBucket", 0)
		mockAPI.AssertNumberOfCalls(t, "DeleteBucket", 0)
	})
	t.Run("New() Should throw error when location id is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
		mockAPI.AssertNumberOfCalls(t, "CreateBucket", 0)
		mockAPI.AssertNumberOfCalls(t, "DeleteBucket", 0)
	})
	t.Run("New() Should throw no error when prefix is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient == nil {
			t.Errorf("New() expected client to not be nil")
			t.Fail()
		}
		if err != nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
		mockAPI.AssertNumberOfCalls(t, "CreateBucket", 0)
		mockAPI.AssertNumberOfCalls(t, "DeleteBucket", 0)
	})
	t.Run("New() Should throw error when service account is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation)

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
		mockAPI.AssertNumberOfCalls(t, "CreateBucket", 0)
		mockAPI.AssertNumberOfCalls(t, "DeleteBucket", 0)
	})
	t.Run("New() Should throw error when no option is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
		mockAPI.AssertNumberOfCalls(t, "CreateBucket", 0)
		mockAPI.AssertNumberOfCalls(t, "DeleteBucket", 0)
	})
	t.Run("New() Should throw error when api is not initialized", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, nil)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Read", 0)
		mockAPI.AssertNumberOfCalls(t, "Write", 0)
		mockAPI.AssertNumberOfCalls(t, "CreateBucket", 0)
		mockAPI.AssertNumberOfCalls(t, "DeleteBucket", 0)
	})
}

func TestClient_CreateBucket(t *testing.T) {
	t.Run("CreateBucket() Should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockAPI.On("CreateBucket", ctx, testGCSBucket).Return(nil)

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		err = mockClient.CreateBucket(ctx, testGCSBucket)
		if err != nil {
			t.Errorf("Client.CreateBucket() error = %v", err)
		}
	})
	t.Run("CreateBucket() Should throw an error when no name is given", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		err = mockClient.CreateBucket(ctx, "")
		if err == nil {
			t.Errorf("Client.CreateBucket() should throw an error")
		}
		mockAPI.AssertNumberOfCalls(t, "CreateBucket", 0)
	})
}

func TestClient_DeleteBucket(t *testing.T) {
	t.Run("DeleteBucket() Should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockAPI.On("DeleteBucket", ctx, testGCSBucket).Return(nil)

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		err = mockClient.DeleteBucket(ctx, testGCSBucket)
		if err != nil {
			t.Errorf("Client.DeleteBucket() error = %v", err)
		}
	})
	t.Run("DeleteBucket() Should throw an error when no name is given", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithPrefix(testGCSPrefix).WithProjectID(testGCSProj).WithLocationID(testGCSLocation).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		err = mockClient.DeleteBucket(ctx, "")
		if err == nil {
			t.Errorf("Client.DeleteBucket() should throw an error")
		}
		mockAPI.AssertNumberOfCalls(t, "DeleteBucket", 0)
	})
}
