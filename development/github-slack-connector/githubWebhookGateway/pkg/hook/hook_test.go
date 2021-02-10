package hook_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/hook"
	"github.com/stretchr/testify/assert"
)

const sampleToken = "1234-567-890"

func exampleHookCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func exampleHookUnprocessableEntity(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnprocessableEntity)
}

func TestCreate(t *testing.T) {
	t.Run("should return nil when response status is equal Created", func(t *testing.T) {
		//given
		handler := http.HandlerFunc(exampleHookCreate)
		server := httptest.NewServer(handler)
		defer server.Close()
		webhook := hook.NewHook("URL")
		//when
		token, err := webhook.Create(sampleToken, server.URL, "secret")
		//then
		assert.NoError(t, err)
		assert.NotEqual(t, "", token)
	})

	t.Run("should return error when response status is not equal Created", func(t *testing.T) {
		//given
		handler := http.HandlerFunc(exampleHookUnprocessableEntity)
		server := httptest.NewServer(handler)
		defer server.Close()
		webhook := hook.NewHook("URL")
		//when
		token, err := webhook.Create(sampleToken, server.URL, "secret")
		//then
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})
}
