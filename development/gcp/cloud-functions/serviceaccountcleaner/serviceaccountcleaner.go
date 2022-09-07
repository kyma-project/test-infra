package serviceaccountcleaner

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	"github.com/kyma-project/test-infra/development/gcp/pkg/iam"
	"github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager"
	gcpiam "google.golang.org/api/iam/v1"
)

var (
	projectID             string
	secretManagerService  *secretmanager.Service
	serviceAccountService *gcpiam.Service
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

	serviceAccountService, err = gcpiam.NewService(ctx)
	if err != nil {
		panic("failed creating IAM client, error: " + err.Error())
	}
}

func ServiceAccountCleaner(w http.ResponseWriter, r *http.Request) {
	var err error

	// Create logger to use google cloud functions structured logging
	logger := cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	logger.WithComponent("kyma.prow.cloud-function.RotateServiceAccount")
	// Set trace value for log entries to identify messages from one function call.
	logger.GenerateTraceValue(projectID, "RotateServiceAccount")

	//options are provided as GET query:
	// time that latest version of secret needs to exist before older ones can be destroyed
	cutoffTimeHours := 5
	keys, ok := r.URL.Query()["age"]
	if ok && len(keys[0]) > 0 {
		cutoffTimeHours, err = strconv.Atoi(keys[0])
		if err != nil {
			w.WriteHeader(400)
			logger.LogCritical("failed to convert age in hours to int: %s", err)
		}
	}

	dryRun := false
	keys, ok = r.URL.Query()["dry_run"]
	if ok && keys[0] == "true" {
		dryRun = true
	}

	//get all secrets that have type=service-account
	projectPath := "projects/" + projectID
	secrets, err := secretManagerService.GetAllSecrets(projectPath, "labels.type=service-account")
	if err != nil {
		logger.LogCritical("Could not get all secrets from %s project", projectPath)
	}

	// for each secret in SA
	for _, secret := range secrets {
		// for each secret in SA:

		//if not excluded
		if secret.Labels["skip-cleanup"] == "true" {
			logger.LogDebug("Secret %s excluded from cleanup, continuing", secret.Labels["type"])
			continue
		}

		// list only enabled versions
		versions, err := secretManagerService.GetAllSecretVersions(secret.Name, "state:enabled")
		if err != nil {
			logger.LogCritical("Could not get enabled versions of a %s secret: %s", secret.Name, err)
		}

		// stop if there are less than two enabled versions
		if len(versions) < 2 {
			logger.LogDebug("less than two enabled versions for %s secret are available, skipping", secret.Name)
			continue
		}

		// if latest older than X seconds
		// timestamp in 2021-07-21T07:31:24.739506Z format
		//2006-01-02T15:04:05
		latestVersionTimestamp, err := time.Parse("2006-01-02T15:04:05", versions[0].CreateTime)
		if err != nil {
			logger.LogCritical("Couldn't parse date: %s", versions[0].CreateTime)
		}
		cutoffTime := time.Now().Add(time.Duration(-cutoffTimeHours) * time.Hour)

		if latestVersionTimestamp.After(cutoffTime) {
			logger.LogInfo("Latest secret is newer than minimal time: %s vs %s", latestVersionTimestamp.String(), cutoffTime.String())
			continue
		}

		for i, version := range versions {
			if i == 0 {
				// skip latest
				continue
			}
			// for each enabled except latest
			versionDataString, err := secretManagerService.GetSecretVersionData(version.Name)
			if err != nil {
				logger.LogCritical("Couldn't get payload of a %s secret: %s", version.Name, err)
			}

			var versionData iam.ServiceAccountJSON
			err = json.Unmarshal([]byte(versionDataString), &versionData)
			if err != nil {
				logger.LogCritical("failed to unmarshal secret JSON field, error: %s", err)
			}

			// get client_email
			serviceAccountKeyPath := "projects/" + versionData.ProjectID + "/serviceAccounts/" + versionData.ClientEmail + "/keys/" + versionData.PrivateKeyID
			logger.LogInfo("Looking for service account %s", serviceAccountKeyPath)

			if !dryRun {
				// delete the key
				keyVersionCall := serviceAccountService.Projects.ServiceAccounts.Keys.Delete(serviceAccountKeyPath)
				_, err = keyVersionCall.Do()
				if err != nil {
					logger.LogError("Could not delete %v key: %s", serviceAccountKeyPath, err)
				}

				// destroy the version
				_, err = secretManagerService.DestroySecretVersion(version.Name)
				if err != nil {
					logger.LogError("Could not destroy %v secret version: %s", version.Name, err)
				}
			} else {
				logger.LogDebug("Dry run: deleting %s and destroying %s", serviceAccountKeyPath, version.Name)
			}
		}

	}
}
