package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/apperrors"

	log "github.com/sirupsen/logrus"
)

// Sender is a struct used to allow mocking the SendToKyma function
type Sender struct {
	validator  Validator
	client     HTTPClient
	serviceURL string
	appName    string
}

// NewSender is a function that creates new Sender with the passed in interfaces
func NewSender(c HTTPClient, v Validator, serviceURL, appName string) Sender {
	return Sender{client: c, validator: v, serviceURL: serviceURL, appName: appName}
}

// HTTPClient is an interface use to allow mocking the http.Client methods
type HTTPClient interface {
	Send(ctx context.Context, event cloudevents.Event) cloudevents.Result
}

// SendToKyma sends the event given by the GitHub API to kyma's event bus
func (k Sender) SendToKyma(eventType, sourceID string, data json.RawMessage) apperrors.AppError {

	t := fmt.Sprintf("sap.kyma.custom.%s.%s.v1", k.appName, eventType)
	kymaEventingType := strings.Replace(t, "-", "", -1)
	log.Info(fmt.Sprintf("publishing event to %s", kymaEventingType))
	event := cloudevents.NewEvent()
	// SourceID is set to value of env variable GITHUB_WEBHOOK_GATEWAY_NAME
	event.SetSource(sourceID)
	event.SetType(kymaEventingType)
	_ = event.SetData(cloudevents.ApplicationJSON, data)

	apperr := k.validator.Validate(event)
	if apperr != nil {
		return apperrors.Internal("while validating the payload: %s", apperr.Error())
	}

	if result := k.client.Send(cloudevents.ContextWithTarget(context.Background(), k.serviceURL), event); cloudevents.IsUndelivered(result) {
		return apperrors.Internal("failed send event to kyma eventing service: %s", result.Error())
	}
	log.Info("sent event to kyma eventing")
	return nil
}
