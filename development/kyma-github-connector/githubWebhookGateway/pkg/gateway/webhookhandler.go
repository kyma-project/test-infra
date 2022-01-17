package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v40/github"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/apperrors"
	git "github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/github"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/httperrors"
	log "github.com/sirupsen/logrus"
)

//Sender is an interface used to allow mocking sending events to Kyma's event bus
type Sender interface {
	SendToKyma(eventType, sourceID string, data json.RawMessage) apperrors.AppError
}

//WebHookHandler is a struct used to allow mocking the github library methods
type WebHookHandler struct {
	validator git.Validator
	sender    Sender
}

//NewWebHookHandler creates a new webhook handler with the passed interface
func NewWebHookHandler(v git.Validator, s Sender) *WebHookHandler {
	return &WebHookHandler{validator: v, sender: s}
}

//HandleWebhook is a function that handles the /webhook endpoint.
func (wh *WebHookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// ValidatePayload will validate request against github webhook spec
	payload, apperr := wh.validator.ValidatePayload(r, []byte(wh.validator.GetToken()))

	if apperr != nil {
		apperr = apperr.Append("while handling '/webhook' endpoint")

		log.Error(apperr.Error())
		httperrors.SendErrorResponse(apperr, w)
		return
	}

	event, apperr := wh.validator.ParseWebHook(github.WebHookType(r), payload)
	if apperr != nil {
		apperr = apperr.Append("While handling '/webhook' endpoint")

		log.Error(apperr.Error())
		httperrors.SendErrorResponse(apperr, w)
		return
	}
	var eventType string
	switch event := event.(type) {
	// Supported github events
	case *github.IssuesEvent:
		eventType = fmt.Sprintf("issuesevent.%s", *event.Action)
	}

	sourceID := os.Getenv("GITHUB_WEBHOOK_GATEWAY_NAME")
	log.Info(fmt.Sprintf("received event of type: %s", eventType))
	apperr = wh.sender.SendToKyma(eventType, sourceID, payload)

	if apperr != nil {
		log.Info(apperrors.Internal("while handling the event: %s", apperr.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}
