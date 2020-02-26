package kubectl

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ()

func TestGenerateKubeconfig(t *testing.T) {
	t.Run("GenerateKubeconfig returns config file path", func(t *testing.T) {
		fakeEndpoint := "1.2.3.4"
		fakeCA := base64.StdEncoding.EncodeToString([]byte("fake CA in base64"))
		fakeName := "fake-name"

		_, err := GenerateKubeconfig(fakeEndpoint, fakeCA, fakeName)
		assert.NoError(t, err)
	})
	t.Run("GenerateKubeconfig should return error if endpoint is not provided", func(t *testing.T) {
		fakeEndpoint := ""
		fakeCA := base64.StdEncoding.EncodeToString([]byte("fake CA in base64"))
		fakeName := "fake-name"

		_, err := GenerateKubeconfig(fakeEndpoint, fakeCA, fakeName)
		assert.EqualError(t, err, "endpoint cannot be empty")
	})
	t.Run("GenerateKubeconfig should return error if fakeCA is not provided", func(t *testing.T) {
		fakeEndpoint := "1.2.3.4"
		fakeCA := ""
		fakeName := "fake-name"

		_, err := GenerateKubeconfig(fakeEndpoint, fakeCA, fakeName)
		assert.EqualError(t, err, "cadata cannot be empty")
	})
	t.Run("GenerateKubeconfig should return error if name is not provided", func(t *testing.T) {
		fakeEndpoint := "1.2.3.4"
		fakeCA := base64.StdEncoding.EncodeToString([]byte("fake CA in base64"))
		fakeName := ""

		_, err := GenerateKubeconfig(fakeEndpoint, fakeCA, fakeName)
		assert.EqualError(t, err, "name cannot be empty")
	})
}
