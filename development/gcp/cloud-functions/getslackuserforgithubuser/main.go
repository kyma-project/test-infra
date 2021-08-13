package getslackuserforgithubuser

import (
"context"
"encoding/json"
"fmt"
"os"

"cloud.google.com/go/functions/metadata"
"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
"github.com/kyma-project/test-infra/development/github/pkg/client"
prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

const (
	TestInfraRepoName = "test-infra"
)

var (
	pubSubClient            *pubsub.Client
	githubClient            *client.Client
	githubAccessToken string
	projectID string
	githubOrg               string
	githubRepo              string
	getSlackUserForCommiterTopic  string
)

func init() {
	var err error
	ctx := context.Background()
	// set variables from environment variables
	projectID = os.Getenv("GCP_PROJECT_ID")
	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	githubOrg = os.Getenv("GITHUB_ORG")
	githubRepo = os.Getenv("GITHUB_REPO")
	getSlackUserForCommiterTopic = os.Getenv("GET_SLACK_USER_FOR_COMMITER_TOPIC")
	// check if variables were set with values
	if getSlackUserForCommiterTopic == "" {
		panic("environment variable GET_SLACK_USER_FOR_COMMITER_TOPIC is empty")
	}
	if projectID == "" {
		panic("environment variable GCP_PROJECT_ID is empty")
	}
	if githubAccessToken == "" {
		panic("environment variable GITHUB_ACCESS_TOKEN is empty")
	}
	if githubOrg == "" {
		panic("environment variable GITHUB_ORG is empty")
	}
	if githubRepo == "" {
		panic("environment variable GITHUB_REPO is empty")
	}
	// create github client
	githubClient, err = client.NewClient(ctx, githubAccessToken)
	if err != nil {
		panic(fmt.Sprintf("Failed creating github client, error: %v", err))
	}
	// create pubsub client
	pubSubClient, err = pubsub.NewClient(ctx, pubsub.PubSubProjectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating pubsub client, error: %v", err))
	}
}


// GetGithubCommiter gets commiter github Login for Refs BaseSHA from pubsub ProwMessage.
// It will find Login for commiter of first Ref in Refs slice because Prow crier pubsub reporter place ProwJobSpec.Refs first.
// ProwJobSpec ExtraRefs are appended second.
func GetGithubCommiter(ctx context.Context, m pubsub.MessagePayload) {
	var err error
	// set trace value to use it in logEntry
	var failingTestMessage pubsub.FailingTestMessage

	// Create logger to use google cloud functions structured logging
	logger := cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	logger.WithComponent("kyma.prow.cloud-function.GetGithubCommiter")
	// Set trace value for log entries to identify messages from one function call.
	logger.GenerateTraceValue(projectID, "GetGithubCommiter")

	// Get metadata from context and set eventID label for logging.
	contextMetadata, err := metadata.FromContext(ctx)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed extract metadata from function call context, error: %s", err.Error()))
	}
	logger.WithLabel("messageId", contextMetadata.EventID)

	// Unmarshall pubsub message data payload.
	err = json.Unmarshal(m.Data, &failingTestMessage)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed unmarshal message data field, error: %s", err.Error()))
	}

	// Set label with prowjob name for logging.
	logger.WithLabel("prowjobName", *failingTestMessage.JobName)

	// Set label with execution ID for logging.
	jobID, err := pubsub.GetJobId(failingTestMessage.URL)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed get job ID, error: %s", err.Error()))
	}
	logger.WithLabel("jobID", *jobID)

	logger.LogInfo(fmt.Sprintf("found prowjob execution ID: %s", *jobID))
