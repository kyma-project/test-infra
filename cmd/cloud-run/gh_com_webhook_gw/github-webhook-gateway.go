// This program receives Github webhook data and sends it as a pubsub message
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/kyma-project/test-infra/pkg/gcp/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/pkg/gcp/http"
	"github.com/kyma-project/test-infra/pkg/gcp/pubsub"

	"github.com/google/go-github/v48/github"
)

var (
	componentName    string
	applicationName  string
	projectID        string
	webhookTokenPath string
	webhookToken     []byte
	listenPort       string
	pubsubClient     *pubsub.Client
)

func main() {
	var err error
	ctx := context.Background()

	componentName = os.Getenv("COMPONENT_NAME")     // github-webhook-gateway
	applicationName = os.Getenv("APPLICATION_NAME") // github-webhook-gateway
	projectID = os.Getenv("PROJECT_ID")
	listenPort = os.Getenv("LISTEN_PORT")
	webhookTokenPath = os.Getenv("WEBHOOK_TOKEN_PATH")

	mainLogger := cloudfunctions.NewLogger()
	mainLogger.WithComponent(componentName) // search-github-issue
	mainLogger.WithLabel("io.kyma.app", applicationName)
	mainLogger.WithLabel("io.kyma.component", componentName)

	webhookToken, err = os.ReadFile(webhookTokenPath)
	if err != nil {
		mainLogger.LogCritical("failed read webhook token from file, error: %s", err)
	}

	pubsubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		mainLogger.LogCritical("Failed creating pubsub client: %s", err)
	}

	http.HandleFunc("/", GithubWebhookGateway)
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

func GithubWebhookGateway(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		payload          []byte
		githubDeliveryID string
	)
	defer r.Body.Close()

	githubDeliveryID = r.Header.Get("X-GitHub-Delivery")

	logger := cloudfunctions.NewLogger()
	logger.WithComponent(componentName)
	logger.WithLabel("io.kyma.app", applicationName)
	logger.WithLabel("io.kyma.component", componentName)

	logger.LogInfo("Got Github payload ID %s from %s", githubDeliveryID, r.URL.Host)

	// payload stores JSON string with webhook data
	payload, err = github.ValidatePayload(r, webhookToken)
	if err != nil {
		// check if wehbook token has beer rotated
		webhookToken, err := os.ReadFile(webhookTokenPath)
		if err != nil {
			logger.LogCritical("failed read github token from file, error: %s", err)
		}
		payload, err = github.ValidatePayload(r, webhookToken)
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed validating Github payload, error: %s", err)
			return
		}
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed parsing Github payload, error: %s", err)
		return
	}

	switch event := event.(type) {
	// Supported github events
	case *github.IssueCommentEvent:
		issueCommentEventRouter(logger, w, event, payload)
	case *github.PullRequestEvent:
		pullRequestEventRouter(logger, w, event, payload)
	case *github.PullRequestReviewCommentEvent:
		pullRequestReviewCommentEventRouter(logger, w, event, payload)
	case *github.PullRequestReviewEvent:
		pullRequestReviewEventRouter(logger, w, event, payload)
	case *github.StatusEvent:
		statusEventRouter(logger, w, event, payload)
	default:
		logger.LogInfo("event %s not supported", github.WebHookType(r))
		w.WriteHeader(http.StatusOK)
	}
}

func issueCommentEventRouter(logger *cloudfunctions.LogEntry, w http.ResponseWriter, event *github.IssueCommentEvent, _ []byte) {
	switch *event.Action {
	case "created":
		publishMessage(logger, w, event, "issue_comment.created")
	case "edited":
		publishMessage(logger, w, event, "issue_comment.edited")
	case "deleted":
		publishMessage(logger, w, event, "issue_comment.deleted")
	default:
		logger.LogInfo("event %s not supported", *event.Action)
		w.WriteHeader(http.StatusOK)
	}
}

func pullRequestEventRouter(logger *cloudfunctions.LogEntry, w http.ResponseWriter, event *github.PullRequestEvent, _ []byte) {
	switch *event.Action {
	case "labeled":
		publishMessage(logger, w, event, "pr.labeled")
	case "unlabeled":
		publishMessage(logger, w, event, "pr.unlabeled")
	case "opened":
		publishMessage(logger, w, event, "pr.opened")
	case "synchronize":
		publishMessage(logger, w, event, "pr.synchronize")
	case "review_requested":
		publishMessage(logger, w, event, "pr.review_requested")
	case "review_dismissed":
		publishMessage(logger, w, event, "pr.review_dismissed")
	default:
		logger.LogInfo("event %s not supported", *event.Action)
		w.WriteHeader(http.StatusOK)
	}
}

func pullRequestReviewCommentEventRouter(logger *cloudfunctions.LogEntry, w http.ResponseWriter, event *github.PullRequestReviewCommentEvent, _ []byte) {
	switch *event.Action {
	case "created":
		publishMessage(logger, w, event, "pull_request_review_comment.created")
	case "edited":
		publishMessage(logger, w, event, "pull_request_review_comment.edited")
	case "deleted":
		publishMessage(logger, w, event, "pull_request_review_comment.deleted")
	default:
		logger.LogInfo("event %s not supported", *event.Action)
		w.WriteHeader(http.StatusOK)
	}
}

func pullRequestReviewEventRouter(logger *cloudfunctions.LogEntry, w http.ResponseWriter, event *github.PullRequestReviewEvent, _ []byte) {
	switch *event.Action {
	case "submitted":
		publishMessage(logger, w, event, "pull_request_review.submitted")
	case "edited":
		publishMessage(logger, w, event, "pull_request_review.edited")
	case "dismissed":
		publishMessage(logger, w, event, "pull_request_review.dismissed")
	default:
		logger.LogInfo("event %s not supported", *event.Action)
		w.WriteHeader(http.StatusOK)
	}
}

// TODO(kacpermalachowski): Consider routing using `state` field.
// See: https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#status
func statusEventRouter(logger *cloudfunctions.LogEntry, w http.ResponseWriter, event *github.StatusEvent, _ []byte) {
	publishMessage(logger, w, event, "status")
	w.WriteHeader(http.StatusOK)
}

func publishMessage(logger *cloudfunctions.LogEntry, w http.ResponseWriter, event interface{}, pubsubTopic string) {
	// send message to a pubsub topic
	ctx := context.Background()
	_, err := pubsubClient.PublishMessage(ctx, event, pubsubTopic)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed sending, error: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
