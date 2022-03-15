package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v40/github"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/apperrors"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/gateway/mocks"
	gitmocks "github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/github/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type toJSON struct {
	TestJSON string `json:"TestJSON"`
	Action   string `json:"Action"`
}

// createRequest creates an HTTP request for test purposes
func createRequest(t *testing.T) *http.Request {

	payload := toJSON{TestJSON: "test", Action: "labeled"}
	toSend, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(toSend))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature", "test")

	return req
}

func TestWebhookHandler(t *testing.T) {
	t.Run("Should respond with 403 status code when given a bad secret", func(t *testing.T) {
		// given

		payload := toJSON{TestJSON: "test"}
		toSend, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(toSend))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mockValidator := &gitmocks.Validator{}
		mockSender := &mocks.Sender{}

		mockValidator.On("ValidatePayload", req).Return(nil, apperrors.AuthenticationFailed("fail"))

		// when
		wh := NewWebHookHandler(mockValidator, mockSender)

		handler := http.HandlerFunc(wh.HandleWebhook)
		handler.ServeHTTP(rr, req)

		// then
		mockValidator.AssertExpectations(t)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

	})

	t.Run("Should respond with 400 status code when given wrong payload ", func(t *testing.T) {

		// given
		req := createRequest(t)
		rr := httptest.NewRecorder()

		mockValidator := &gitmocks.Validator{}
		mockSender := &mocks.Sender{}
		mockPayload, err := json.Marshal(toJSON{TestJSON: "test"})
		require.NoError(t, err)

		mockValidator.On("ValidatePayload", req, []byte("test")).Return(mockPayload, nil)
		mockValidator.On("ParseWebHook", "", mockPayload).Return(nil, apperrors.WrongInput("fail"))

		wh := NewWebHookHandler(mockValidator, mockSender)

		// when
		handler := http.HandlerFunc(wh.HandleWebhook)
		handler.ServeHTTP(rr, req)

		// then
		mockValidator.AssertExpectations(t)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Should respond with 200 status code, when given a payload with a known event", func(t *testing.T) {

		// given
		req := createRequest(t)
		req.Header.Set("X-Github-Event", "issues")
		rr := httptest.NewRecorder()

		mockValidator := &gitmocks.Validator{}
		mockSender := &mocks.Sender{}
		mockPayload, err := json.Marshal(toJSON{TestJSON: "test", Action: "labeled"})
		require.NoError(t, err)
		rawPayload := json.RawMessage(mockPayload)
		mockSender.On("SendToKyma", "issuesevent.labeled", "", rawPayload).Return(nil)

		mockValidator.On("ValidatePayload", req, []byte("test")).Return(mockPayload, nil)
		var action = "labeled"
		event := &github.IssuesEvent{Action: &action}
		mockValidator.On("ParseWebHook", "issues", mockPayload).Return(event, nil)

		wh := NewWebHookHandler(mockValidator, mockSender)

		// when
		handler := http.HandlerFunc(wh.HandleWebhook)
		handler.ServeHTTP(rr, req)

		// then
		mockValidator.AssertExpectations(t)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Should respond with 400 status code, when given a payload with an unknown event", func(t *testing.T) {

		// given
		req := createRequest(t)
		rr := httptest.NewRecorder()

		mockValidator := &gitmocks.Validator{}
		mockSender := &mocks.Sender{}

		mockPayload, err := json.Marshal(toJSON{TestJSON: "test"})
		require.NoError(t, err)
		mockValidator.On("ValidatePayload", req, []byte("test")).Return(mockPayload, nil)
		mockValidator.On("ParseWebHook", "", mockPayload).Return(nil, apperrors.NotFound("Unknown event"))
		wh := NewWebHookHandler(mockValidator, mockSender)

		// when
		handler := http.HandlerFunc(wh.HandleWebhook)
		handler.ServeHTTP(rr, req)

		// then
		mockValidator.AssertExpectations(t)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

}
