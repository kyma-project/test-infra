package events

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/apperrors"
)

// validator is a struct used to allow mocking the validatePayload function
type validator struct {
	Validator
}

// Validator is an interface used to allow mocking the validatePayload function
type Validator interface {
	Validate(payload cloudevents.Event) apperrors.AppError
}

// NewValidator returns new Validator
func NewValidator() Validator {
	return validator{}
}

// Validate method checks the given payload fields
func (v validator) Validate(event cloudevents.Event) apperrors.AppError {

	if event.Type() == "" {
		return apperrors.WrongInput("cloudevent type should not be empty")
	}
	if event.Source() == "" {
		return apperrors.WrongInput("cloudevent source should not be empty")
	}
	if len(event.Data()) == 0 {
		return apperrors.WrongInput("cloudevent data should not be empty")
	}

	return nil
}
