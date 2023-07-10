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
	gcptypes "github.com/kyma-project/test-infra/development/gcp/pkg/types"
	kgithubv1 "github.com/kyma-project/test-infra/development/github/pkg/client"
	githubtypes "github.com/kyma-project/test-infra/development/github/pkg/types"
	"github.com/kyma-project/test-infra/development/types"
	"golang.org/x/net/context"
)

var (
	componentName        string
	applicationName      string
	projectID            string
	listenPort           string
	toolsGithubTokenPath string
	githubOrg            string // "neighbors-team"
	githubRepo           string // "leaks-test"
	sapGhClient          *kgithubv1.SapToolsClient
	gcsClient            *storage.Client
)

type message struct {
	pubsub.ProwMessage
	types.SecretsLeakScannerMessage
	githubtypes.SearchIssuesResult
	gcptypes.GCPBucketMetadata
	gcptypes.GCPProjectMetadata
	githubtypes.IssueMetadata
}

func main() {
	var (
		err error
	)
	componentName = os.Getenv("COMPONENT_NAME")     // issue-creator
	applicationName = os.Getenv("APPLICATION_NAME") // github-bot
	projectID = os.Getenv("PROJECT_ID")
	listenPort = os.Getenv("LISTEN_PORT")
	githubOrg = os.Getenv("GITHUB_ORG")
	githubRepo = os.Getenv("GITHUB_REPO")
	toolsGithubTokenPath = os.Getenv("TOOLS_GITHUB_TOKEN_PATH")

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

	githubToken, err := os.ReadFile(toolsGithubTokenPath)
	if err != nil {
		mainLogger.LogCritical("failed read github token from file, error: %s", err)
	}

	sapGhClient, err = kgithubv1.NewSapToolsClient(ctx, string(githubToken))
	if err != nil {
		mainLogger.LogCritical("failed create sap github client")
	}

	http.HandleFunc("/", searchGithubIssues)
	// Determine listenPort for HTTP service.
	if listenPort == "" {
		listenPort = "8080"
		mainLogger.LogInfo("Defaulting to listenPort %s", listenPort)
	}
	// Start HTTP server.
	mainLogger.LogInfo("Listening on listenPort %s", listenPort)
	if err := http.ListenAndServe(":"+listenPort, nil); err != nil {
		mainLogger.LogError("failed listen on listenPort %s, error: %s", listenPort, err)
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
		logger.LogError("failed dump http request, error: %s", err)
	}
	logger.LogDebug("request:\n%v", string(requestDump))

	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusBadRequest, logger, "failed to parse CloudEvent from request: %s", err.Error())
		return
	}

	logger.LogInfo("got message, id: %s, type: %s", event.ID(), event.Type())

	// Load event data
	if err = event.DataAs(&msg); err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed marshal event, error: %s", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hasher := sha1.New()
	hasher.Write([]byte(*msg.JobName + *msg.JobType))
	h := hasher.Sum(nil)
	secretsleakscannerID := base64.StdEncoding.EncodeToString(h)

	// TODO: potentially a query and other parameters could be passed in a call to the service, so the service can be used by multiple tools.
	// 	This should increase a caching efficiency when implemented.
	query := fmt.Sprintf("secretsleakscanner_id=%s in:body org:%s repo:%s is:issue is:open", secretsleakscannerID, githubOrg, githubRepo)

	// Search issues
	opts := &github.SearchOptions{
		Sort:        "",
		Order:       "",
		TextMatch:   false,
		ListOptions: github.ListOptions{},
	}

	var (
		searchResult *github.IssuesSearchResult
		result       *github.Response
	)
	sapGhClient.WrapperClientMu.RLock()
	searchResult, result, err = sapGhClient.Search.Issues(ctx, query, opts)
	sapGhClient.WrapperClientMu.RUnlock()
	if result != nil {
		switch {
		case result.StatusCode == 401:
			logger.LogWarning("Github authentication failed, got %d response status code, trying to reauthenticate", result.StatusCode)
			githubToken, err := os.ReadFile(toolsGithubTokenPath)
			if err != nil {
				logger.LogCritical("failed read github token from file, error: %s", err)
			}
			_, err = sapGhClient.Reauthenticate(ctx, logger, githubToken)
			if err != nil {
				logger.LogCritical("failed reauthenticate github client, error %s", err)
			}
			// Retry GitHub API call with eventually new credentials. This may fail again because of no new credentials provided.
			sapGhClient.WrapperClientMu.RLock()
			searchResult, result, err = sapGhClient.Search.Issues(ctx, query, opts)
			sapGhClient.WrapperClientMu.RUnlock()
			if result != nil && (result.StatusCode < 200 || result.StatusCode >= 300) {
				crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed search github issues, received non 2xx response code, error: %s", err)
				return
			}
		case result.StatusCode < 200 || result.StatusCode >= 300:
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed search github issues, received non 2xx response code, error: %s", err)
			return
		}
	}
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed search github issues, error: %s", err)
		return
	}

	issues := searchResult.Issues
	responseEvent := cloudevents.NewEvent()
	responseEvent.SetSource(applicationName + "/" + componentName)
	responseEvent.SetID(applicationName + "/" + componentName + "/" + trace)
	if len(issues) != 0 {
		msg.GithubIssueFound = github.Bool(true)
		msg.GithubIssues = issues
		logger.LogInfo("found github issues")
		responseEvent.SetType("sap.tools.github.leakissue.found")
		if err = responseEvent.SetData(cloudevents.ApplicationJSON, msg); err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed set event data, error: %s", err)
			return
		}
	} else {
		logger.LogInfo("github issues not found")
		responseEvent.SetType("sap.tools.github.leakissue.notfound")
		msg.GithubIssueFound = github.Bool(false)
		if err = responseEvent.SetData(cloudevents.ApplicationJSON, msg); err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed set event data, error: %s", err)
			return
		}
	}
	headers := w.Header()
	headers.Set("Content-Type", cloudevents.ApplicationJSON)
	headers.Set("X-Cloud-Trace-Context", traceHeader)
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(responseEvent); err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed write response body, error: %s", err.Error())
		return
	}
}
