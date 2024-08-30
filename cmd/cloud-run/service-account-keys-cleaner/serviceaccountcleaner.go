package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-project/test-infra/pkg/gcp/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/pkg/gcp/http"
	"github.com/kyma-project/test-infra/pkg/gcp/iam"
	"github.com/kyma-project/test-infra/pkg/gcp/secretmanager"

	"cloud.google.com/go/compute/metadata"
	gcpiam "google.golang.org/api/iam/v1"
)

var (
	secretManagerService  *secretmanager.Service
	serviceAccountService *gcpiam.Service
	componentName         string
	applicationName       string
	projectID             string
	listenPort            string
)

func main() {
	var err error

	componentName = os.Getenv("COMPONENT_NAME")     // issue-creator
	applicationName = os.Getenv("APPLICATION_NAME") // secrets-rotator
	listenPort = os.Getenv("LISTEN_PORT")

	mainLogger := cloudfunctions.NewLogger()
	mainLogger.WithComponent(componentName)
	mainLogger.WithLabel("io.kyma.app", applicationName)
	mainLogger.WithLabel("io.kyma.component", componentName)

	ctx := context.Background()

	projectID, err = metadata.ProjectIDWithContext(ctx)
	if err != nil {
		mainLogger.LogCritical("failed to retrieve GCP Project ID, error: %s", err.Error())
	}

	secretManagerService, err = secretmanager.NewService(ctx)
	if err != nil {
		mainLogger.LogCritical("failed creating Secret Manager client, error: %s", err.Error())
	}

	serviceAccountService, err = gcpiam.NewService(ctx)
	if err != nil {
		mainLogger.LogCritical("failed creating IAM client, error: %s", err.Error())
	}

	http.HandleFunc("/", serviceAccountKeysCleaner)
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

// serviceAccountCleaner destroys old versions of service-account secrets and corresponding keys
func serviceAccountKeysCleaner(w http.ResponseWriter, r *http.Request) {
	var (
		trace       string
		traceHeader string
		err         error
	)

	// set trace value to use it in logEntry
	traceHeader = r.Header.Get("X-Cloud-Trace-Context")

	if projectID != "" {
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace = fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[0])
		}
	}

	// Create logger to use Google structured logging
	logger := cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	logger.WithComponent(componentName)
	logger.WithLabel("io.kyma.app", applicationName)
	logger.WithLabel("io.kyma.component", componentName)
	logger.WithTrace(trace)

	// options are provided as GET query:
	// time that latest version of secret needs to exist before older ones can be destroyed
	cutoffTimeHours := 1
	keys, ok := r.URL.Query()["age"]
	if ok && len(keys[0]) > 0 {
		cutoffTimeHours, err = strconv.Atoi(keys[0])
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusBadRequest, logger, "failed to convert age in hours to int: %s", err)
			return
		}
	}

	dryRun := false
	keys, ok = r.URL.Query()["dry_run"]
	if ok && keys[0] == "true" {
		dryRun = true
	}

	keys, ok = r.URL.Query()["project"]
	if ok {
		projectID = keys[0]
	}

	// get all secrets that have type=service-account
	projectPath := "projects/" + projectID
	secrets, err := secretManagerService.GetAllSecrets(projectPath, "labels.type=service-account")
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed to get all secrets from %s project", projectPath)
		return
	}

	// for each secret in SA
	for _, secret := range secrets {
		// for each secret in SA:

		// if not excluded
		if secret.Labels["skip-cleanup"] == "true" {
			logger.LogDebug("Secret %s excluded from cleanup, continuing", secret.Labels["type"])
			continue
		}

		// list only enabled versions
		versions, err := secretManagerService.GetAllSecretVersions(secret.Name, "state:enabled")
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "Could not get enabled versions of a %s secret: %s", secret.Name, err)
			return
		}

		// stop if there are less than two enabled versions
		if len(versions) < 2 {
			logger.LogDebug("less than two enabled versions for %s secret are available, skipping", secret.Name)
			continue
		}

		// if latest older than X seconds
		// timestamp in 2021-07-21T07:31:24.739506Z format
		// 2006-01-02T15:04:05
		latestVersionTimestamp, err := time.Parse("2006-01-02T15:04:05.000000Z", versions[0].CreateTime)
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "Couldn't parse date: %s", versions[0].CreateTime)
			return
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
				crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "Couldn't get payload of a %s secret: %s", version.Name, err)
				return
			}

			var versionData iam.ServiceAccountJSON
			err = json.Unmarshal([]byte(versionDataString), &versionData)
			if err != nil {
				crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed to unmarshal secret JSON field, error: %s", err)
				return
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
	w.WriteHeader(http.StatusOK)
}
