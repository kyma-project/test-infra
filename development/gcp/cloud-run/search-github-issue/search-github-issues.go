package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/v42/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/development/gcp/pkg/http"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	kgithubv1 "github.com/kyma-project/test-infra/development/github/pkg/client"
	kgithub "github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"github.com/kyma-project/test-infra/development/types"
	"golang.org/x/net/context"
)

var (
	componentName   string
	applicationName string
	projectID       string
	// bucketName       string
	port        string
	githubOrg   string // "neighbors-team"
	githubRepo  string // "leaks-test"
	githubToken []byte
	// githubSecretPath string
	sapGhClient *kgithubv1.SapToolsClient
	gcsClient   *storage.Client
)

type message struct {
	pubsub.ProwMessage
	types.SecretsLeakScannerMessage
	kgithub.SearchIssuesResult
	types.GCPStorageMetadata
	types.GCPProjectMetadata
	types.GithubIssueMetadata
}

func main() {
	var (
		err error
	)
	componentName = os.Getenv("COMPONENT_NAME")     // issue-creator
	applicationName = os.Getenv("APPLICATION_NAME") // github-bot
	projectID = os.Getenv("PROJECT_ID")
	port = os.Getenv("LISTEN_PORT")
	githubOrg = os.Getenv("GITHUB_ORG")
	githubRepo = os.Getenv("GITHUB_REPO")

	mainLogger := cloudfunctions.NewLogger()
	mainLogger.WithComponent(componentName) // search-github-issue
	mainLogger.WithLabel("io.kyma.app", applicationName)
	mainLogger.WithLabel("io.kyma.component", componentName)

	ctx := context.Background()

	gcsClient, err = storage.NewClient(ctx)
	if err != nil {
		mainLogger.LogCritical("failed to create client: %s", err.Error())
	}
	defer gcsClient.Close()

	githubToken, err = os.ReadFile("/etc/github-token/github-token")

	sapGhClient, err = kgithubv1.NewSapToolsClient(ctx, string(githubToken))
	if err != nil {
		mainLogger.LogCritical("failed create sap github client")
	}

	http.HandleFunc("/", searchGithubIssues)
	// Determine port for HTTP service.
	if port == "" {
		port = "8080"
		mainLogger.LogInfo("Defaulting to port %s", port)
	}
	// Start HTTP server.
	mainLogger.LogInfo("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		mainLogger.LogError(err.Error())
	}
}

func searchGithubIssues(w http.ResponseWriter, r *http.Request) {
	var (
		msg         message
		trace       string
		traceHeader string
	)

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

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logger.LogError("failed dump http request, error:", err)
	}
	logger.LogDebug("request:\n%v", string(requestDump))

	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusBadRequest, logger, "failed to parse CloudEvent from request: %s", err.Error())
		return
	}

	logger.LogInfo("got message, id: %s, type: %s", event.ID(), event.Type())

	// Load event data
	if err = event.DataAs(&msg); err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed marshal event, error: %s", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hasher := sha1.New()
	hasher.Write([]byte(*msg.JobName + *msg.JobType))
	h := hasher.Sum(nil)
	secretsleakscannerID := base64.StdEncoding.EncodeToString(h)

	// TODO: potentialy a query and other parameters could be passed in a call to the service, so the service can be used by multiple tools.
	// 	This should increase a caching efficiency when implemented.
	query := fmt.Sprintf("secretsleakscanner_id=%s in:body org:%s repo:%s is:issue is:open", secretsleakscannerID, githubOrg, githubRepo)

	// Search issues
	opts := &github.SearchOptions{
		Sort:        "",
		Order:       "",
		TextMatch:   false,
		ListOptions: github.ListOptions{},
	}
	searchResult, result, err := sapGhClient.Search.Issues(ctx, query, opts)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed search github issues, error: %s", err)
		return
	}

	_, err = kgithubv1.IsStatusOK(result)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed search github issues, error: %s", err)
		return
	}

	issues := searchResult.Issues
	responseEvent := cloudevents.NewEvent()
	responseEvent.SetSource(applicationName + "/" + componentName)
	responseEvent.SetID(applicationName + "/" + componentName + "/" + trace)
	if len(issues) != 0 {
		msg.IssueFound = github.Bool(true)
		msg.Issues = issues
		logger.LogInfo("found github issues")
		responseEvent.SetType("sap.tools.github.leakissue.found")
		if err = responseEvent.SetData(cloudevents.ApplicationJSON, msg); err != nil {
			crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed set event data, error: %s", err)
			return
		}
		// body, err = json.Marshal(responseEvent)
		// if err != nil {
		// 	crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed marshal event, error: %s", err.Error())
		// 	return
		// }
	} else {
		logger.LogInfo("github issues not found")
		responseEvent.SetType("sap.tools.github.leakissue.notfound")
		msg.IssueFound = github.Bool(false)
		if err = responseEvent.SetData(cloudevents.ApplicationJSON, msg); err != nil {
			crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed set event data, error: %s", err)
			return
		}
		// body, err = json.Marshal(responseEvent)
		// if err != nil {
		// 	crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed marshal event, error: %s", err.Error())
		// 	return
		// }
	}
	headers := w.Header()
	headers.Set("Content-Type", cloudevents.ApplicationJSON)
	headers.Set("X-Cloud-Trace-Context", traceHeader)
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(responseEvent); err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed write response body, error: %s", err.Error())
		return
	}
	// w.Write(body)
}
