package github

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/apperrors"
)

type eventsReceiver struct {
	secret string
}

//Validator is an interface used to allow mocking the github library methods
type Validator interface {
	ValidatePayload(*http.Request, []byte) ([]byte, apperrors.AppError)
	ParseWebHook(string, []byte) (interface{}, apperrors.AppError)
	GetToken() string
}

//NewReceivingEventsWrapper return receivingEventsWrapper struct
func NewReceivingEventsWrapper(s string) Validator {
	return &eventsReceiver{secret: s}
}

//ValidatePayload is a function used for checking whether the secret provided in the request is correct
func (wh eventsReceiver) ValidatePayload(r *http.Request, b []byte) ([]byte, apperrors.AppError) {
	payload, err := github.ValidatePayload(r, b)
	if err != nil {
		return nil, apperrors.AuthenticationFailed("authentication during GitHub payload validation failed: %s", err)
	}
	return payload, nil
}

//ParseWebHook parses the raw json payload into an event struct
func (wh eventsReceiver) ParseWebHook(s string, b []byte) (interface{}, apperrors.AppError) {
	webhook, err := github.ParseWebHook(s, b)
	if err != nil {
		return nil, apperrors.WrongInput("failed to parse incomming github payload into struct: %s", err)
	}
	return webhook, nil
}

//GetToken is a function that looks for the secret in the environment
func (wh eventsReceiver) GetToken() string {
	return wh.secret
}
