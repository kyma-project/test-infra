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

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/v42/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/development/gcp/pkg/http"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	gcptypes "github.com/kyma-project/test-infra/development/gcp/pkg/types"
	kgithubv1 "github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/github/pkg/templates"
	githubtypes "github.com/kyma-project/test-infra/development/github/pkg/types"
	"github.com/kyma-project/test-infra/development/types"
	"golang.org/x/net/context"
)

var (
	componentName   string
	applicationName string
	projectID       string
	githubToken     []byte
	githubOrg       string // "neighbors-team"
	githubRepo      string // "leaks-test"
	port            string
	sapGhClient     *kgithubv1.SapToolsClient
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
	var err error
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

	githubToken, err = os.ReadFile("/etc/github-token/github-token")
	if err != nil {
		mainLogger.LogCritical("failed read github token from file, error: %s", err)
	}

	sapGhClient, err = kgithubv1.NewSapToolsClient(ctx, string(githubToken))
	if err != nil {
		mainLogger.LogCritical("failed create sap github client")
	}

	http.HandleFunc("/", createGithubIssue)
	// Determine port for HTTP service.
	if port == "" {
		port = "8080"
		mainLogger.LogInfo("Defaulting to port %s", port)
	}
	// Start HTTP server.
	mainLogger.LogInfo("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		mainLogger.LogCritical("failed listen on port %s, error: %s", port, err)
	}
}

func createGithubIssue(w http.ResponseWriter, r *http.Request) {
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
		logger.LogError("failed dump http request, error: %s", err.Error())
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

	issueData := templates.SecretsLeakIssueData{
		SecretsLeaksScannerID:     secretsleakscannerID,
		ProwMessage:               msg.ProwMessage,
		SecretsLeakScannerMessage: msg.SecretsLeakScannerMessage,
	}

	issueBody, err := issueData.RenderBody()
	if err != nil {
		logger.LogError("failed render issue body, error: %s", err.Error())
	}

	issueTitle := fmt.Sprintf("Secret leak found in %s prowjob logs", *msg.JobName)
	issueLabels := []string{"bug"}
	issueState := github.String("open")
	issueRequest := github.IssueRequest{
		Title:     github.String(issueTitle),
		Body:      github.String(fmt.Sprint(issueBody.String())),
		Labels:    &issueLabels,
		Assignee:  github.String(msg.GithubIssueAssignee.SapToolsGithubUsername),
		State:     issueState,
		Milestone: nil,
		Assignees: nil,
	}
	// Search issues
	issue, response, err := sapGhClient.Issues.Create(ctx, githubOrg, githubRepo, &issueRequest)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed create github issues, error: %s", err)
		return
	}
	ok, err := kgithubv1.IsStatusOK(response)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed create github issues, failed read status code, error %s", err)
		return
	}
	if !ok {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed create github issues, received non 2xx response code: status code %d", response.StatusCode)
		return
	}
	logger.LogInfo("created github issue: %s", issue.GetHTMLURL())
	msg.GithubIssueNumber = issue.Number
	msg.GithubIssueURL = issue.HTMLURL
	msg.GithubIssueOrg = github.String(githubOrg)
	msg.GithubIssueRepo = github.String(githubRepo)
	// TODO: Add setting GithubAssigne in msg

	// process issue
	responseEvent := cloudevents.NewEvent()
	responseEvent.SetSource(applicationName + "/" + componentName)
	responseEvent.SetID(applicationName + "/" + componentName + "/" + trace)
	responseEvent.SetType("sap.tools.github.issue.created")
	if err = responseEvent.SetData(cloudevents.ApplicationJSON, msg); err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed set event data, error: %s", err)
		return
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
