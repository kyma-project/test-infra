package events_test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/apperrors"
	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/events"
	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/events/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type toJSON struct {
	TestJSON string `json:"TestJSON"`
}

type ClientMock struct {
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	if req.URL.String() == "test" {
		return nil, apperrors.Internal("Couldn't create a request")
	}
	return &http.Response{StatusCode: 200}, nil
}

func TestSendToKyma(t *testing.T) {
	t.Run("should return no error when given proper arguments", func(t *testing.T) {
		k := events.NewSender(&ClientMock{}, events.NewValidator(), "http://event-bus-publish.kyma-system:8080/v1/events")
		payload := toJSON{TestJSON: "test"}
		toSend, err := json.Marshal(payload)
		require.NoError(t, err)
		assert.Equal(t, nil, k.SendToKyma("issuesevent.labeled", "github-connector-app", "v1", "", toSend))
	})

	t.Run("should return an internal error when wrong arguments", func(t *testing.T) {
		//given
		payload := toJSON{TestJSON: "test"}
		toSend, err := json.Marshal(payload)
		require.NoError(t, err)
		mockValidator := &mocks.Validator{}
		mockValidator.On("Validate", events.EventRequestPayload{EventTypeVersion: "v1",
			EventTime: time.Now().Format(time.RFC3339),
			SourceID:  "github-connector-app",
			Data:      toSend}).Return(apperrors.Internal("test"))

		k := events.NewSender(&ClientMock{}, mockValidator, "http://event-bus-publish.kyma-system:8080/v1/events")
		expected := apperrors.Internal("test")

		//when
		actual := k.SendToKyma("", "github-connector-app", "v1", "", toSend)

		//then
		assert.Equal(t, expected.Code(), actual.Code())
	})

	t.Run("should return an internal error when couldn't send a request", func(t *testing.T) {
		//given
		payload := toJSON{TestJSON: "test"}
		toSend, err := json.Marshal(payload)
		require.NoError(t, err)
		mockValidator := &mocks.Validator{}
		mockValidator.On("Validate", events.EventRequestPayload{EventTypeVersion: "v1",
			EventTime: time.Now().Format(time.RFC3339),
			SourceID:  "github-connector-app",
			Data:      toSend}).Return(apperrors.Internal("test"))

		k := events.NewSender(&ClientMock{}, mockValidator, "test")
		expected := apperrors.Internal("test")

		//when
		actual := k.SendToKyma("", "github-connector-app", "v1", "", toSend)

		//then
		assert.Equal(t, expected.Code(), actual.Code())
	})

	t.Run("should return no error when server responded with a 200 status code", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			checkEventRequest(t, r)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		validatorMock := &mocks.Validator{}
		payload := toJSON{TestJSON: "test"}
		toSend, err := json.Marshal(payload)
		require.NoError(t, err)

		validatorMock.On("Validate", events.EventRequestPayload{EventType: "issuesevent.labeled", EventTypeVersion: "v1", EventTime: time.Now().Format(time.RFC3339), SourceID: "github-connector-app", Data: toSend}).Return(nil)
		sender := events.NewSender(&http.Client{}, events.NewValidator(), ts.URL)

		// when
		apperr := sender.SendToKyma("issuesevent.labeled", "github-connector-app", "v1", "", toSend)

		// then
		require.NoError(t, apperr)

	})

	t.Run("should return an error when server didn't respond with a 200 status code", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			checkEventRequest(t, r)
			w.WriteHeader(http.StatusInternalServerError)
		}))

		defer ts.Close()

		validatorMock := &mocks.Validator{}
		payload := toJSON{TestJSON: "test"}
		toSend, err := json.Marshal(payload)
		require.NoError(t, err)

		validatorMock.On("Validate", events.EventRequestPayload{EventTime: time.Now().Format(time.RFC3339), SourceID: "github-connector-app", Data: toSend}).Return(nil)
		sender := events.NewSender(&http.Client{}, events.NewValidator(), ts.URL)

		// when
		apperr := sender.SendToKyma("", "github-connector-app", "", "", toSend)

		// then
		require.Error(t, apperr)
		log.Println(apperr.Code())
		assert.Equal(t, true, apperr.Code() == apperrors.CodeInternal)
	})

}

func checkEventRequest(t *testing.T, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	testStruct := events.EventRequestPayload{}
	err := decoder.Decode(&testStruct)
	require.NoError(t, err)

	assert.Equal(t, "issuesevent.labeled", testStruct.EventType)
	assert.Equal(t, "v1", testStruct.EventTypeVersion)
	assert.Equal(t, "github-connector-app", testStruct.SourceID)
}
