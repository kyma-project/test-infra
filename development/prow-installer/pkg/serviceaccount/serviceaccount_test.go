package serviceaccount

import (
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("NewClient() succeed.", func(t *testing.T) {
		mockIAM := &mocks.IAM{}
		prefix := "test_prefix"
		client := NewClient(prefix, mockIAM)
		assert.Equal(t, prefix, client.prefix)
		assert.Equal(t, mockIAM, client.iamservice)
		assert.NotNilf(t, client.iamservice, "")
	})
	t.Run("NewClient() fail because passed nil value for iamservice.", func(t *testing.T) {

	})
}
