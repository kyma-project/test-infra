package secrets

import (
	"context"
	"testing"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/secrets/automock"

	"github.com/stretchr/testify/assert"
)

var (
	testKmsProj     = "kms-test-project"
	testKmsLocation = "kms-test-location"
	testKmsRing     = "kms-test-ring"
	testKmsKey      = "kms-test-key"

	testSecretVal    = "hello-this-is-secret"
	testEncryptedVal = "encrypted-secret"
)

func TestClient_Encrypt(t *testing.T) {
	t.Run("Encrypt() Should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockAPI.On("Encrypt", ctx, []byte(testSecretVal)).Return([]byte(testEncryptedVal), nil)

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		data, err := mockClient.Encrypt(ctx, []byte(testSecretVal))
		if err != nil {
			t.Errorf("Client.Encrypt() error = %v", err)
		}
		assert.Equal(t, string(data), testEncryptedVal)
	})
	t.Run("Encrypt() Should throw error when len(0) byte is passed", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		data, err := mockClient.Encrypt(ctx, []byte(""))
		if err == nil {
			t.Errorf("Client.Encrypt() should have thrown an error")
		}
		assert.Nil(t, data)
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
	})
}

func TestClient_Decrypt(t *testing.T) {
	t.Run("Decrypt() Should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockAPI.On("Decrypt", ctx, []byte(testEncryptedVal)).Return([]byte(testSecretVal), nil)

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		data, err := mockClient.Decrypt(ctx, []byte(testEncryptedVal))
		if err != nil {
			t.Errorf("Client.Decrypt() error = %v", err)
		}
		assert.Equal(t, string(data), testSecretVal)
	})
	t.Run("Decrypt() Should throw an error when len(0) byte is passed", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		ctx := context.Background()
		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if err != nil {
			t.Errorf("failed before running a test")
			t.Fail()
		}

		data, err := mockClient.Decrypt(ctx, []byte(""))
		if err == nil {
			t.Errorf("Client.Decrypt() should have thrown an error")
		}
		assert.Nil(t, data)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
	})
}

func TestNew(t *testing.T) {
	t.Run("New() should not throw errors", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient == nil {
			t.Errorf("New() expected client to not be nil")
			t.Fail()
		}
		if err != nil {
			t.Errorf("New() error should be nil %v", err)
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
	})
	t.Run("New() Should throw error when project id is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
	})
	t.Run("New() Should throw error when location id is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
	})
	t.Run("New() Should throw error when kms ring is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
	})
	t.Run("New() Should throw error when kms key is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
	})
	t.Run("New() Should throw error when service account is not present", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey)

		mockClient, err := New(opts, mockAPI)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
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
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
	})
	t.Run("New() Should throw error when api is not initialized", func(t *testing.T) {
		mockAPI := &automock.API{}
		defer mockAPI.AssertExpectations(t)

		opts := Option{}
		opts = opts.WithProjectID(testKmsProj).WithLocationID(testKmsLocation).WithKmsRing(testKmsRing).WithKmsKey(testKmsKey).WithServiceAccount("not-empty-gcp-will-validate")

		mockClient, err := New(opts, nil)
		if mockClient != nil {
			t.Errorf("New() expected client to be nil")
			t.Fail()
		}
		if err == nil {
			t.Errorf("New() error is nil, expected an error")
			t.Fail()
		}
		mockAPI.AssertNumberOfCalls(t, "Encrypt", 0)
		mockAPI.AssertNumberOfCalls(t, "Decrypt", 0)
	})
}
