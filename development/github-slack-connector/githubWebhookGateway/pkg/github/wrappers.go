package github

import (
	"net/http"

	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/apperrors"
)

type receivingEventsWrapper struct {
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
	return &receivingEventsWrapper{secret: s}
}

//ValidatePayload is a function used for checking whether the secret provided in the request is correct
func (wh receivingEventsWrapper) ValidatePayload(r *http.Request, b []byte) ([]byte, apperrors.AppError) {
	payload, err := github.ValidatePayload(r, b)
	if err != nil {
		return nil, apperrors.AuthenticationFailed("Authentication during GitHub payload validation failed: %s", err)
	}
	return payload, nil
}

//ParseWebHook parses the raw json payload into an event struct
func (wh receivingEventsWrapper) ParseWebHook(s string, b []byte) (interface{}, apperrors.AppError) {
	webhook, err := github.ParseWebHook(s, b)
	if err != nil {
		return nil, apperrors.WrongInput("Failed to parse incomming github payload into struct: %s", err)
	}
	return webhook, nil
}

//GetToken is a function that looks for the secret in the environment
func (wh receivingEventsWrapper) GetToken() string {
	return wh.secret
}
