package rotateserviceaccount

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"cloud.google.com/go/compute/metadata"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager"
	"google.golang.org/api/iam/v1"
)

var (
	projectID             string
	secretManagerService  *secretmanager.Service
	serviceAccountService *iam.Service
)

// ServiceAccountJSON stores Service Account athentication data
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

	projectID, err = metadata.ProjectID()
	if err != nil {
		panic("failed to retrieve GCP Project ID, error: " + err.Error())
	}

	secretManagerService, err = secretmanager.NewService(ctx)
	if err != nil {
		panic("failed creating Secret Manager client, error: " + err.Error())
	}

	serviceAccountService, err = iam.NewService(ctx)
	if err != nil {
		panic("failed creating IAM client, error: " + err.Error())
	}
}

// RotateServiceAccount adds new secret version in Secret Manager on pubsub rotate message
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
		logger.LogDebug("Unsupported event type: " + m.Attributes["eventType"] + ", quitting")
		return nil
	}

	err = json.Unmarshal(m.Data, &secretRotateMessage)
	if err != nil {
		logger.LogCritical("failed to unmarshal message data field, error: " + err.Error())
	}

	if secretRotateMessage.Labels["type"] != "service-account" {
		logger.LogDebug("Unsupported secret type: " + secretRotateMessage.Labels["type"] + ", quitting")
		return nil
	}

	//get latest secret version data
	secretlatestVersionPath := secretRotateMessage.Name + "/versions/latest"
	logger.LogInfo("Retrieving secret: " + secretlatestVersionPath)
	secretDataString, err := secretManagerService.GetSecretVersionData(secretlatestVersionPath)
	if err != nil {
		logger.LogCritical("failed to retrieve latest version of a secret " + secretRotateMessage.Name + ", error: " + err.Error())
	}

	logger.LogInfo("Trying to unmarshal secret: " + secretRotateMessage.Name)
	decodedSecretDataString, err := base64.StdEncoding.DecodeString(secretDataString)
	if err != nil {
		logger.LogCritical("Could not base64 decode secret " + secretRotateMessage.Name)
	}
	err = json.Unmarshal([]byte(decodedSecretDataString), &secretData)
	if err != nil {
		logger.LogCritical("failed to unmarshal secret JSON field, error: " + err.Error())
	}

	// get client_email
	serviceAccountPath := "projects/" + secretData.ProjectID + "/serviceAccounts/" + secretData.ClientEmail
	logger.LogInfo("Looking for service account" + serviceAccountPath)
	createKeyRequest := iam.CreateServiceAccountKeyRequest{}
	newKeyCall := serviceAccountService.Projects.ServiceAccounts.Keys.Create(serviceAccountPath, &createKeyRequest)
	newKey, err := newKeyCall.Do()
	if err != nil {
		logger.LogCritical("failed to create new key for " + serviceAccountPath + " Service Account, error: " + err.Error())
	}

	logger.LogInfo("Decoding new key data for " + serviceAccountPath)
	newKeyBytes, err := base64.StdEncoding.DecodeString(newKey.PrivateKeyData)
	if err != nil {
		logger.LogCritical("failed to decode new key for " + serviceAccountPath + " Service Account, error: " + err.Error())
	}

	// update secret
	logger.LogInfo("Adding new secret version to secret " + secretRotateMessage.Name)
	_, err = secretManagerService.AddSecretVersion(secretRotateMessage.Name, newKeyBytes)
	if err != nil {
		logger.LogCritical("failed to create new " + secretRotateMessage.Name + " secret version, error: " + err.Error())
	}

	return nil
}
