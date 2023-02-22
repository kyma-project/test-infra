package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	gpubsub "cloud.google.com/go/pubsub"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/development/gcp/pkg/http"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager"
	"google.golang.org/api/iam/v1"
)

var (
	secretManagerService  *secretmanager.Service
	serviceAccountService *iam.Service
	componentName         string
	applicationName       string
	projectID             string
	listenPort            string
)

// ServiceAccountJSON stores Service Account authentication data
type ServiceAccountJSON struct {
	Type             string `json:"type"`
	ProjectID        string `json:"project_id"`
	PrivateKeyID     string `json:"private_key_id"`
	PrivateKey       string `json:"private_key"`
	ClientEmail      string `json:"client_email"`
	ClientID         string `json:"client_id"`
	AuthURL          string `json:"auth_uri"`
	TokenURI         string `json:"token_uri"`
	AuthProviderCert string `json:"auth_provider_x509_cert_url"`
	ClientCert       string `json:"client_x509_cert_url"`
}

func main() {
	var err error

	componentName = os.Getenv("COMPONENT_NAME")     // issue-creator
	applicationName = os.Getenv("APPLICATION_NAME") // github-bot
	listenPort = os.Getenv("LISTEN_PORT")

	mainLogger := cloudfunctions.NewLogger()
	mainLogger.WithComponent(componentName) // search-github-issue
	mainLogger.WithLabel("io.kyma.app", applicationName)
	mainLogger.WithLabel("io.kyma.component", componentName)

	ctx := context.Background()

	projectID, err = metadata.ProjectID()
	if err != nil {
		mainLogger.LogCritical("failed to retrieve GCP Project ID, error: " + err.Error())
	}

	secretManagerService, err = secretmanager.NewService(ctx)
	if err != nil {
		mainLogger.LogCritical("failed creating Secret Manager client, error: " + err.Error())
	}

	serviceAccountService, err = iam.NewService(ctx)
	if err != nil {
		mainLogger.LogCritical("failed creating IAM client, error: " + err.Error())
	}

	http.HandleFunc("/", rotateServiceAccount)
	// Determine listenPort for HTTP service.
	if listenPort == "" {
		listenPort = "8080"
		mainLogger.LogInfo("Defaulting to listenPort %s", listenPort)
	}
	// Start HTTP server.
	mainLogger.LogInfo("Listening on listenPort %s", listenPort)
	if err := http.ListenAndServe(":"+listenPort, nil); err != nil {
		mainLogger.LogCritical("failed listen on listenPort %s, error: %s", listenPort, err)
	}
}

// rotateServiceAccount adds new secret version in Secret Manager on pubsub rotate message
func rotateServiceAccount(w http.ResponseWriter, r *http.Request) {
	var (
		trace               string
		traceHeader         string
		secretData          ServiceAccountJSON
		secretRotateMessage pubsub.SecretRotateMessage
		err                 error
	)

	// set trace value to use it in logEntry
	traceHeader = r.Header.Get("X-Cloud-Trace-Context")

	if projectID != "" {
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace = fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[0])
		}
	}

	logger := cloudfunctions.NewLogger()
	logger.WithComponent(componentName)
	logger.WithLabel("io.kyma.app", applicationName)
	logger.WithLabel("io.kyma.component", componentName)
	logger.WithTrace(trace)

	// decode http messages body
	var message gpubsub.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed decode message body")
		return
	}

	logger.WithLabel("messageId", message.ID)

	if message.Attributes["eventType"] != "SECRET_ROTATE" {
		logger.LogDebug("Unsupported event type: %s, quitting", message.Attributes["eventType"])
		w.WriteHeader(http.StatusOK)
		return
	}

	err = json.Unmarshal(message.Data, &secretRotateMessage)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusBadRequest, logger, "failed to unmarshal message data field, error: %s", err.Error())
		return
	}

	if secretRotateMessage.Labels["type"] != "service-account" {
		logger.LogDebug("Unsupported secret type: %s, quitting", secretRotateMessage.Labels["type"])
		w.WriteHeader(http.StatusOK)
		return
	}

	// get latest secret version data
	logger.LogInfo("Retrieving latest version of secret: %s", secretRotateMessage.Name)
	secretDataString, err := secretManagerService.GetLatestSecretVersionData(secretRotateMessage.Name)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed to retrieve latest version of a secret %s, error: %s", secretRotateMessage.Name, err.Error())
		return
	}

	err = json.Unmarshal([]byte(secretDataString), &secretData)
	if err != nil {
		logger.LogCritical("failed to unmarshal secret JSON field, error: %s", err.Error())
	}

	// get client_email
	serviceAccountPath := "projects/" + secretData.ProjectID + "/serviceAccounts/" + secretData.ClientEmail
	logger.LogInfo("Looking for service account %s", serviceAccountPath)
	createKeyRequest := iam.CreateServiceAccountKeyRequest{}
	newKeyCall := serviceAccountService.Projects.ServiceAccounts.Keys.Create(serviceAccountPath, &createKeyRequest)
	newKey, err := newKeyCall.Do()
	if err != nil {
		logger.LogCritical("failed to create new key for %s Service Account, error: %s", serviceAccountPath, err.Error())
	}

	logger.LogInfo("Decoding new key data for %s", serviceAccountPath)
	newKeyBytes, err := base64.StdEncoding.DecodeString(newKey.PrivateKeyData)
	if err != nil {
		logger.LogCritical("failed to decode new key for %s Service Account, error: %s", serviceAccountPath, err.Error())
	}

	// update secret
	logger.LogInfo("Adding new secret version to secret %s", secretRotateMessage.Name)
	_, err = secretManagerService.AddSecretVersion(secretRotateMessage.Name, newKeyBytes)
	if err != nil {
		logger.LogCritical("failed to create new %s secret version, error: %s", secretRotateMessage.Name, err.Error())
	}

	w.WriteHeader(http.StatusOK)
}
