package events

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/apperrors"

	log "github.com/sirupsen/logrus"
)

//Sender is a struct used to allow mocking the SendToKyma function
type Sender struct {
	validator  Validator
	client     HTTPClient
	serviceURL string
}

//NewSender is a function that creates new Sender with the passed in interfaces
func NewSender(c HTTPClient, v Validator, serviceURL string) Sender {
	return Sender{client: c, validator: v, serviceURL: serviceURL}
}

//HTTPClient is an interface use to allow mocking the http.Client methods
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// EventRequestPayload represents a POST request's body which is sent to Event-Service
type EventRequestPayload struct {
	EventType        string          `json:"type"`
	EventTypeVersion string          `json:"specversion"`
	EventID          string          `json:"id,omitempty"` //uuid should be generated automatically if send empty
	EventTime        string          `json:"time"`
	SourceID         string          `json:"source"`         //put your application name here
	Data             json.RawMessage `json:"data,omitempty"` //github webhook json payload
}

//SendToKyma is a function that sends the event given by the GitHub API to kyma's event bus
//func (k Sender) SendToKyma(eventType, eventTypeVersion, eventID, sourceID string, data json.RawMessage) apperrors.AppError {
func (k Sender) SendToKyma(eventType, sourceID, eventTypeVersion, eventID string, data json.RawMessage) apperrors.AppError {

	payload := EventRequestPayload{
		eventType,
		eventTypeVersion,
		eventID,
		time.Now().Format(time.RFC3339),
		sourceID,
		data}

	apperr := k.validator.Validate(payload)
	if apperr != nil {
		return apperrors.Internal("While validating the payload: %s", apperr.Error())
	}

	jsonToSend, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return apperrors.Internal("Can not marshall given struct: %s", err.Error())
	}

	kymaRequest, err := http.NewRequest(http.MethodPost, k.serviceURL,
		bytes.NewReader(jsonToSend))
	if err != nil {
		return apperrors.Internal("While creating an http request: %s", err.Error())
	}

	response, err := k.client.Do(kymaRequest)
	if err != nil {
		return apperrors.Internal("While sending the event to the EventBus: %s", err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return apperrors.Internal("Error sending event: %d", response.StatusCode)
	}

	log.Info(response)
	return nil
}
