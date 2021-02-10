package events

import (
	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/apperrors"
)

// validator is a struct used to allow mocking the validatePayload function
type validator struct {
	Validator
}

//Validator is an interface used to allow mocking the validatePayload function
type Validator interface {
	Validate(payload EventRequestPayload) apperrors.AppError
}

//NewValidator is a function that creates a validator struct with the passed in interface
func NewValidator() validator {
	return validator{}
}

// Validate method checks the given payload fields
func (v validator) Validate(payload EventRequestPayload) apperrors.AppError {

	if payload.EventType == "" {
		return apperrors.WrongInput("eventType should not be empty")
	}
	if payload.EventTypeVersion == "" {
		return apperrors.WrongInput("eventTypeVersion should not be empty")
	}
	if payload.SourceID == "" {
		return apperrors.WrongInput("sourceID should not be empty")
	}
	if len(payload.Data) == 0 {
		return apperrors.WrongInput("data should not be empty")
	}

	return nil
}
