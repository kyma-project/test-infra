// This program receives Github webhook data and sends it as a pubsub message
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/test-infra/pkg/gcp/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/pkg/gcp/http"
	"github.com/kyma-project/test-infra/pkg/gcp/pubsub"
	toolsclient "github.com/kyma-project/test-infra/pkg/github/client"
	"github.com/kyma-project/test-infra/pkg/types"
	"net/http"
	"os"

	"github.com/google/go-github/v48/github"
)

var (
	// TODO: allowedEvents map should be populated from configuration.
	//  This will allow to limit allowed events by instance, event when code support it.
	// Event types allowed processing by this instance.
	allowedEvents = map[string]map[string]struct{}{
		"issuesevent": {
			"labeled": struct{}{},
		},
	}

	componentName        string
	applicationName      string
	projectID            string
	toolsGithubTokenPath string
	githubToken          []byte
	webhookTokenPath     string
	webhookToken         []byte
	pubsubTopic          string
	listenPort           string
	sapToolsClient       *toolsclient.SapToolsClient
	pubsubClient         *pubsub.Client
)

func main() {
	var err error
	ctx := context.Background()

	componentName = os.Getenv("COMPONENT_NAME")     // github-webhook-gateway
	applicationName = os.Getenv("APPLICATION_NAME") // github-webhook-gateway
	projectID = os.Getenv("PROJECT_ID")
	listenPort = os.Getenv("LISTEN_PORT")
	pubsubTopic = os.Getenv("PUBSUB_TOPIC")
	toolsGithubTokenPath = os.Getenv("TOOLS_GITHUB_TOKEN_PATH")
	webhookTokenPath = os.Getenv("WEBHOOK_TOKEN_PATH")

	mainLogger := cloudfunctions.NewLogger()
	mainLogger.WithComponent(componentName) // search-github-issue
	mainLogger.WithLabel("io.kyma.app", applicationName)
	mainLogger.WithLabel("io.kyma.component", componentName)

	githubToken, err = os.ReadFile(toolsGithubTokenPath)
	if err != nil {
		mainLogger.LogCritical("failed read github token from file, error: %s", err)
	}

	webhookToken, err = os.ReadFile(webhookTokenPath)
	if err != nil {
		mainLogger.LogCritical("failed read webhook token from file, error: %s", err)
	}

	// Create tools github client.
	sapToolsClient, err = toolsclient.NewSapToolsClient(ctx, string(githubToken))
	if err != nil {
		mainLogger.LogCritical("Failed creating tools GitHub client: %s", err)
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
		eventType        string
		supported        bool
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
	case *github.IssuesEvent:
		eventType, supported = checkIfEventSupported(allowedEvents, "issuesevent", *event.Action)
	default:
		supported = false
	}
	if supported {
		var usersMap []types.User
		ctx := context.Background()
		sapToolsClient.WrapperClientMu.RLock()
		usersMap, err = sapToolsClient.GetUsersMap(ctx)
		sapToolsClient.WrapperClientMu.RUnlock()
		if err != nil {
			githubToken, err := os.ReadFile(toolsGithubTokenPath)
			if err != nil {
				logger.LogCritical("failed read github token from file, error: %s", err)
			}
			_, err = sapToolsClient.Reauthenticate(ctx, logger, githubToken)
			if err != nil {
				logger.LogCritical("failed reauthenticate github client, error %s", err)
			}

			// retry
			sapToolsClient.WrapperClientMu.RLock()
			usersMap, err = sapToolsClient.GetUsersMap(ctx)
			sapToolsClient.WrapperClientMu.RUnlock()
			if err != nil {
				crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed getting user map, error: %s", err)
				return
			}
		}
		// CloudEvents sourceID.
		logger.LogInfo("received event of type: %s", eventType)
		issue := event.(*github.IssuesEvent).GetIssue()
		sender := event.(*github.IssuesEvent).GetSender()

		// add Slack user name, or empty string
		var payloadInterface map[string]any
		json.Unmarshal(payload, &payloadInterface)

		if issue.Assignee != nil {
			// assigneee can be null
			assigneeSlackUsername := getSlackUsername(usersMap, *issue.Assignee.Login)
			payloadInterface["assigneeSlackUsername"] = assigneeSlackUsername
		} else {
			payloadInterface["assigneeSlackUsername"] = ""
		}

		senderSlackUsername := getSlackUsername(usersMap, *sender.Login)
		payloadInterface["senderSlackUsername"] = senderSlackUsername

		// send message to a pubsub topic
		_, err = pubsubClient.PublishMessage(ctx, payloadInterface, pubsubTopic)
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed sending, error: %s", err)
			return
		}
	} else {
		logger.LogInfo("received unsupported event")
	}
	w.WriteHeader(http.StatusOK)
}

// checkIfEventSupported will check if eventGroup and eventAction are present in allowed map of allowed event types for this instance.
// If group and action is allowed, function will return event type.
func checkIfEventSupported(allowed map[string]map[string]struct{}, eventGroup, eventAction string) (string, bool) {
	// Check if event type is allowed
	if _, ok := allowed[eventGroup][eventAction]; ok {
		et := fmt.Sprintf("%s.%s", eventGroup, eventAction)
		return et, true
	}
	return "", false
}

// getSlackusername loks through usersmap and returns GH username
func getSlackUsername(usersMap []types.User, githubUsername string) string {
	for _, user := range usersMap {
		if githubUsername == user.SapToolsGithubUsername {
			return user.ComEnterpriseSlackUsername
		}
	}

	return ""
}
# (2025-03-04)