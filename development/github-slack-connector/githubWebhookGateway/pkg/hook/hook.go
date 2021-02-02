package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/apperrors"
)

const (
	kymaURLPrefix = "https://"
	kymaURLSuffix = "/webhook"
	kymaURLFormat = "%s%s%s"
)

//Hook describe hook struct
type Hook interface {
	Create(t string, githubURL string, secret string) (string, apperrors.AppError)
}

//Hook is a struct that contains information about github's repo/org url, OAuth token and allows creating webhooks
type hook struct {
	kymaURL string
}

//NewHook create Hook structure
func NewHook(URL string) Hook {
	kURL := fmt.Sprintf(kymaURLFormat, kymaURLPrefix, URL, kymaURLSuffix)
	return &hook{kymaURL: kURL}
}

//Create build request and create webhook in github's repository or organization
func (s *hook) Create(t string, githubURL string, secret string) (string, apperrors.AppError) {
	token := fmt.Sprintf("token %s", t)
	hook := PayloadDetails{
		Name:   "web",
		Active: true,
		Config: Config{
			URL:         s.kymaURL,
			InsecureSSL: "1",
			ContentType: "json",
			Secret:      secret,
		},
		Events: []string{"*"},
	}

	payloadJSON, err := json.Marshal(hook)
	if err != nil {
		return "", apperrors.Internal("Failed to marshal hook: %s", err.Error())
	}

	requestReader := bytes.NewReader(payloadJSON)
	httpRequest, err := http.NewRequest(http.MethodPost, githubURL, requestReader)

	if err != nil {
		return "", apperrors.Internal("Failed to create JSON request: %s", err.Error())
	}

	httpRequest.Header.Set("Authorization", token)

	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return "", apperrors.UpstreamServerCallFailed("Failed to make request to '%s': %s", githubURL, err.Error())
	}

	if httpResponse.StatusCode != http.StatusCreated {
		return "", apperrors.UpstreamServerCallFailed("Unpredicted response code: %v", httpResponse.StatusCode)
	}
	return secret, nil
}
