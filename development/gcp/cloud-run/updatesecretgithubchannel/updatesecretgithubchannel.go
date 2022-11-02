package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"cloud.google.com/go/compute/metadata"
	// "github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/google/go-github/v42/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/github/pkg/client"
)

// const (
// TODO: check if projectID can be read from some env vars.
// projectID = "sap-kyma-prow"
// )

var (
	projectID         string
	err               error
	githubAccessToken string
	githubClient      *client.SapToolsClient
	logger            *cloudfunctions.LogEntry
)

// TODO: remove secret from yaml key names, this data is redundant. The file itself describe secret.
// Alias holds mapping between owners file alias and slack groups and channels names.
// It holds information if automerge notification is enabled.
type SyncEvent struct {
	SecretName      string   `yaml:"secret.name,omitempty"`
	SecretGCPPath   string   `yaml:"gcpPath"`
	SecretVersion   int      `yaml:"secret.version,omitempty"`
	SecretEndpoints []string `yaml:"secret.endpoints,omitempty"`
}

func main() {
	var err error
	ctx := context.Background()

	logger = cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	_ = logger.WithComponent("kyma.prow.cloud-function.UpdateSecretGitHubChannel")

	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")

	projectID, err = metadata.ProjectID()
	if err != nil {
		panic("failed to retrieve GCP Project ID, error: " + err.Error())
	}

	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubAccessToken == "" {
		panic("environment variable GITHUB_ACCESS_TOKEN is empty")
	}
	// create github client
	githubClient, err = client.NewSapToolsClient(ctx, githubAccessToken)
	if err != nil {
		logger.LogCritical("Failed creating github client, error: %w", err)
	}

	http.HandleFunc("/", myHTTPFunction)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		logger.LogInfo("Defaulting to port %s", port)
	}
	// Start HTTP server.
	logger.LogInfo("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.LogCritical(err.Error())
	}
}

// Function myHTTPFunction is an HTTP handler
func myHTTPFunction(w http.ResponseWriter, r *http.Request) {
	// Your code here
	ctx := context.Background()
	// if projectID == "" {
	//	projectID, err = metadata.ProjectID()
	//	if err != nil {
	//		panic("failed to retrieve GCP Project ID, error: " + err.Error())
	//	}
	// }

	// Create logger to use google cloud functions structured logging
	//	logger := cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	//	logger.WithComponent("kyma.prow.cloud-function.UpdateSecretGitHubChannel")
	// Set trace value for log entries to identify messages from one function call.
	logger.LogInfo(projectID)
	logger.GenerateTraceValue(projectID, "UpdateSecretGitHubChannel")

	var message pubsub.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		logger.LogError("failed decode message body")
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: Use w.Write function instead fmt.Fprint
		if _, err := fmt.Fprint(w, "500 - failed decode message!"); err != nil {
			logger.LogError("failed send HTTP response")
		}
		return
	}

	logger.WithLabel("messageId", message.Message.MessageID)

	if message.Message.Data == nil {
		logger.LogError("message data is empty, nothing to analyse")
		w.WriteHeader(http.StatusBadRequest)
		if _, err := fmt.Fprint(w, "400 - message is empty!"); err != nil {
			logger.LogError("failed send response for message id: %s", message.Message.MessageID)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("NOK"))
		}
	}

	if path.Base(message.Message.Attributes["secretId"]) != "test-secret-service-account-sap-kyma-prow" {
		logger.LogInfo("unsupported secret: %s", path.Base(message.Message.Attributes["secretId"]))
		io.WriteString(w, "OK")
	}

	// got valid message
	logger.LogInfo("received message with id: %s", message.Message.MessageID)

	// decode base64 prow message
	// bdata := make([]byte, base64.StdEncoding.DecodedLen(len(message.Message.Data)))
	// _, err := base64.StdEncoding.Decode(bdata, message.Message.Data)
	// if err != nil {
	//	logger.LogError("prow message data field base64 decoding failed, error: %w", err)
	//	w.WriteHeader(http.StatusInternalServerError)
	//	_, _ = w.Write([]byte("NOK"))
	//	return
	// }
	logger.LogInfo(string(message.Message.Data))
	logger.LogInfo("%+v", message.Message.Attributes)

	if message.Message.Attributes["eventType"] == "SECRET_VERSION_ADD" {

		var syncEvent SyncEvent
		syncEventFilePath := "/" + message.Message.Attributes["secretId"] + ".yaml"
		// Get file from github.
		logger.LogInfo("getting sync event file from gtihub.tools.sap%s", syncEventFilePath)
		syncEventFile, _, resp, err := githubClient.Repositories.GetContents(ctx, "kyma", "test-infra", syncEventFilePath, &github.RepositoryContentGetOptions{Ref: "secrets-sync"})
		if err != nil {
			logger.LogError("got error when getting sync event file from github.tools.sap, error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("NOK"))
			return
		}
		// Check HTTP response code
		if ok, err := githubClient.IsStatusOK(resp); !ok {
			logger.LogError("Response status for getting file content from github is not OK, error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("NOK"))
			return
		}
		// Read file content.
		syncEventString, err := syncEventFile.GetContent()
		if err != nil {
			logger.LogError("got error when getting content of sync event file, error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("NOK"))
			return
		}
		err = yaml.Unmarshal([]byte(syncEventString), &syncEvent)
		if err != nil {
			logger.LogError("got error when unmarshaling sync event file content, error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("NOK"))
			return
		}
		versionString := path.Base(message.Message.Attributes["versionId"])
		logger.LogInfo("secret versionId to update: %s", versionString)
		versionInt, err := strconv.Atoi(versionString)
		if err != nil {
			logger.LogError("Failed convert secret version to sync from string to integer, error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("NOK"))
			return
		}
		syncEvent.SecretVersion = versionInt
		syncEvent.SecretGCPPath = message.Message.Attributes["secretId"]
		updatedSyncEventFile, err := yaml.Marshal(syncEvent)
		opts := &github.RepositoryContentFileOptions{
			Message:   github.String("New secret version added. Synchronization needed."),
			Content:   updatedSyncEventFile,
			SHA:       syncEventFile.SHA,
			Branch:    github.String("secrets-sync"),
			Committer: &github.CommitAuthor{Name: github.String("SecretSync Bot"), Email: github.String("user@example.com")},
		}
		_, syncEventFileGitHubPath, found := strings.Cut(syncEventFilePath, "/")
		if !found {
			logger.LogError("failed remove leading slash from event sync file path, provided file path: %s", syncEventFilePath)
		}
		reposContentResponse, response, err := githubClient.Repositories.UpdateFile(ctx, "kyma", "test-infra", syncEventFileGitHubPath, opts)
		if err != nil {
			logger.LogError("got error when updating sync event file in github.tools.sap, error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("NOK"))
			return
		}
		if ok, err := githubClient.IsStatusOK(response); !ok {
			logger.LogError("Response status for updating sync event file in github is not OK, error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("NOK"))
			return
		}
		logger.LogInfo("Updated sync event file %s in github.tools.sap with commit %s", syncEventFilePath, reposContentResponse.Commit.SHA)
	}
	// Send an HTTP response
	io.WriteString(w, "OK")
}
