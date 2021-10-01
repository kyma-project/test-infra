package getgithubcommiter

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
	pubSubClient                 *pubsub.Client
	githubClient                 *client.Client
	githubAccessToken            string
	projectID                    string
	githubOrg                    string
	githubRepo                   string
	getSlackUserForCommiterTopic string
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
	pubSubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating pubsub client, error: %v", err))
	}
}

// GetGithubCommiter gets commiter github Login for Refs BaseSHA from pubsub ProwMessage.
// It will find Login for commiter of first Ref in Refs slice because Prow crier pubsub reporter place ProwJobSpec.Refs first.
// ProwJobSpec ExtraRefs are appended second.
func GetGithubCommiter(ctx context.Context, m pubsub.MessagePayload) error {
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

	// Find all SHA
	var commitersLogins []string
	var refs []prowapi.Refs
	// Check if working with periodic type prowjob.
	// For periodic jobs Refs has only ProwJobSpec.ExtraRefs. If more than one Refs present in slice,
	// we need to exclude test-infra Refs as it probably deliver only tools to test another repo.
	if pjType := failingTestMessage.JobType; *pjType == string(prowapi.PeriodicJob) {
		// Check if periodic prowjob Refs holds only one Ref.
		// Add it to the refs slice to get commiter login.
		if len(failingTestMessage.Refs) == 1 {
			refs = append(refs, failingTestMessage.Refs[0])
			// Find non test-infra Refs when there is more than one in a Refs slice.
		} else {
			// Add all valid Refs to refs slice.
			for _, ref := range failingTestMessage.Refs {
				if ref.Repo != TestInfraRepoName {
					refs = append(refs, ref)
				}
			}
		}
		// For non periodic prowjobs the first Ref in Refs slice is the one against which prowjob was running.
	} else {
		refs = append(refs, failingTestMessage.Refs[0])
	}
	// Finding github commiters Logins for all collected Refs.
	for _, ref := range refs {
		// Find commiter Login.
		comiterLogin, err := githubClient.GetAuthorLoginForSHA(ctx, ref.BaseSHA, ref.Org, ref.Repo)
		if err != nil {
			logger.LogCritical(fmt.Sprintf("failed find commiter login for %s/%s commit SHA %s, got error %v", githubOrg, githubRepo, failingTestMessage.Refs[0].BaseSHA, err))
		}
		// Add commiter Login to the slice with all Logins.
		commitersLogins = append(commitersLogins, *comiterLogin)
	}
	// Add all Logins to the pubsub failing test message.
	// This message will be publish to the pubsub as a pubsub message payload.
	failingTestMessage.GithubCommitersLogins = commitersLogins
	// call pubsub to add commiter to firestore
	// cal pubsub to add comment on github issue "Commiter XYZ requested to check this issue".
	// Publish message for further processing.
	publlishedMessageID, err := pubSubClient.PublishMessage(ctx, failingTestMessage, getSlackUserForCommiterTopic)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()))
	}
	logger.LogInfo(fmt.Sprintf("published pubsub message to topic %s, id: %s", getSlackUserForCommiterTopic, *publlishedMessageID))
	return nil
}
