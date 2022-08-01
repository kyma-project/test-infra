package rotateserviceaccount

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager"
	"github.com/kyma-project/test-infra/development/gcp/pkg/secretversionsmanager"
	"google.golang.org/api/iam/v1"
)

var (
	projectID                   string
	secretManagerService        *secretmanager.Service
	secretVersionManagerService *secretversionsmanager.Service
	serviceAccountService       *iam.Service
)

type ServiceAccountJSON struct {
	Type             string `json:"type"`
	ProjectID        string `json:"project_id"`
	PrivatekayID     string `json:"private_key_id"`
	PrivateKey       string `json:"private_key"`
	ClientEmail      string `json:"client_email"`
	ClientID         string `json:"client_id"`
	AuthURL          string `json:"auth_uri"`
	TokenURI         string `json:"token_uri"`
	AuthProviderCert string `json:"auth_provider_x509_cert_url"`
	ClientCert       string `json:"client_x509_cert_url"`
}

func init() {
	var err error
	ctx := context.Background()

	projectID = os.Getenv("GCP_PROJECT_ID")

	secretManagerService, err = secretmanager.NewService(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed creating Secret Manager client, error: %s", err.Error()))
	}
	secretVersionManagerService = secretversionsmanager.NewService(secretManagerService)

	serviceAccountService, err = iam.NewService(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed creating IAM client, error: %s", err.Error()))
	}
}

func RotateServiceAccount(ctx context.Context, m pubsub.MessagePayload) error {
	var err error
	var secretRotateMessage pubsub.SecretRotateMessage
	var secretData ServiceAccountJSON

	// Create logger to use google cloud functions structured logging
	logger := cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	logger.WithComponent("kyma.prow.cloud-function.RotateServiceAccount")
	// Set trace value for log entries to identify messages from one function call.
	logger.GenerateTraceValue(projectID, "RotateServiceAccount")

	if m.Attributes["eventType"] != "SECRET_ROTATE" {
		logger.LogInfo(fmt.Sprintf("Unsupported event type: %s, quitting\n", m.Attributes["eventType"]))
		return nil
	}

	err = json.Unmarshal(m.Data, &secretRotateMessage)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed to unmarshal message data field, error: %s", err.Error()))
	}

	//get latest secret version data
	secretlatestVersionPath := secretRotateMessage.Name + "/versions/latest"
	logger.LogInfo(fmt.Sprintf("Retrieving %s secret", secretlatestVersionPath))
	secretDataString, err := secretVersionManagerService.GetSecretVersionData(secretlatestVersionPath)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed to retreive latest version of a secret %s, error: %s", secretRotateMessage.Name, err.Error()))
	}

	logger.LogInfo(fmt.Sprintf("Trying to unmarshal %s secret", secretRotateMessage.Name))
	decodedSecretDataString, err := base64.StdEncoding.DecodeString(secretDataString)
	err = json.Unmarshal([]byte(decodedSecretDataString), &secretData)
	if err != nil {
		logger.LogError(fmt.Sprintf("failed to unmarshal secret JSON field, error: %s", err.Error()))
		logger.LogInfo("The secret might not store a Service Account credentials")
		return nil
	}

	// get client_email
	serviceAccountPath := "projects/" + secretData.ProjectID + "/serviceAccounts/" + secretData.ClientEmail
	logger.LogInfo(fmt.Sprintf("Looking for %s service account", serviceAccountPath))
	createKeyRequest := iam.CreateServiceAccountKeyRequest{}
	newKeyCall := serviceAccountService.Projects.ServiceAccounts.Keys.Create(serviceAccountPath, &createKeyRequest)
	newKey, err := newKeyCall.Do()
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed to create new key for %s Service Account, error: %s", serviceAccountPath, err.Error()))
	}

	logger.LogInfo(fmt.Sprintf("Decoding %s new key data", serviceAccountPath))
	newKeyBytes, err := base64.StdEncoding.DecodeString(newKey.PrivateKeyData)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed to decode new key for %s Service Account, error: %s", serviceAccountPath, err.Error()))
	}

	// update secret
	logger.LogInfo(fmt.Sprintf("Adding new %s secret version", secretRotateMessage.Name))
	_, err = secretManagerService.AddSecretVersion(secretRotateMessage.Name, newKeyBytes)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed to create new %s secret version, error: %s", secretRotateMessage.Name, err.Error()))
	}

	return nil
}
