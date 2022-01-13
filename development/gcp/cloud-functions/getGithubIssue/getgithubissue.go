package getGithubIssue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/pubsub"
	"github.com/google/go-github/v40/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	kymapubsub "github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"golang.org/x/oauth2"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

const (
	TestInfraRepoName = "test-infra"
)

var (
	firestoreClient         *firestore.Client
	pubSubClient            *pubsub.Client
	githubClient            *github.Client
	ts                      oauth2.TokenSource
	projectID               string
	githubAccessToken       string
	firestoreCollection     string
	getGithubCommiterTopic  string
	getProwjobErrorsTopic   string
	getFailureInstanceTopic string
)

func init() {
	var err error
	ctx := context.Background()
	// set variables from environment variables
	projectID = os.Getenv("GCP_PROJECT_ID")
	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	firestoreCollection = os.Getenv("FIRESTORE_COLLECTION")
	getGithubCommiterTopic = os.Getenv("GET_GITHUB_COMMITER_TOPIC")
	getProwjobErrorsTopic = os.Getenv("GET_PROWJOB_ERRORS_TOPIC")
	getFailureInstanceTopic = os.Getenv("GET_FAILURE_INSTANCE_TOPIC")
	// check if variables were set with values
	if getGithubCommiterTopic == "" {
		panic("environment variable GET_GITHUB_COMMITER_TOPIC is empty")
	}
	if getProwjobErrorsTopic == "" {
		panic("environment variable GET_PROWJOB_ERRORS_TOPIC is empty")
	}
	if getFailureInstanceTopic == "" {
		panic("environment variable GET_FAILURE_INSTANCE_TOPIC is empty")
	}
	if projectID == "" {
		panic("environment variable GCP_PROJECT_ID is empty")
	}
	if githubAccessToken == "" {
		panic("environment variable GITHUB_ACCESS_TOKEN is empty")
	}
	if firestoreCollection == "" {
		panic("environment variable FIRESTORE_COLLECTION is empty, can't setup firebase client")
	}
	// create firestore client, it will be reused by multiple function calls
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating firestore client, error: %s", err.Error()))
	}
	// create pubsub client, it will be reused by multiple function calls
	pubSubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating pubsub client, error: %s", err.Error()))
	}
	// create github client with user token authentication, it will be reused by multiple function calls
	ts = oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: githubAccessToken,
			TokenType:   "token",
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient = github.NewClient(tc)
}

// isGithubIssueOpen checks if github issue from pubsub message data payload is open.
func isGithubIssueOpen(ctx context.Context, client *github.Client, message kymapubsub.FailingTestMessage) (*bool, *github.Issue, error) {
	// Call github API to get required issue.
	ghIssue, ghResponse, err := client.Issues.Get(ctx, *message.GithubIssueOrg, *message.GithubIssueRepo, int(*message.GithubIssueNumber))
	// Check API response and err, if issue was not found or API call failed.
	if ghResponse != nil {
		err = github.CheckResponse(ghResponse.Response)
		if err != nil {
			return nil, nil, fmt.Errorf("github API call reply with error, error: %w", err)
		}
	} else if err != nil {
		return nil, nil, fmt.Errorf("calling github API failed, error: %w", err)
	}
	b := new(bool)
	if *ghIssue.State == "open" {
		// Set return value to true when issue is open.
		b = github.Bool(true)
	} else {
		// Set return value to false when issue is closed
		b = github.Bool(false)
	}
	return b, ghIssue, nil
}

// TODO: creating github issues should be moved to separate cloud function.
// createGithubIssue will create new github issue with data from pubsub message data payload.
func createGithubIssue(ctx context.Context, client *github.Client, message kymapubsub.FailingTestMessage) (*github.Issue, error) {
	var ref prowapi.Refs
	// Check if working with periodic type prowjob.
	// For periodic jobs Refs has only ProwJobSpec.ExtraRefs. If more than one Refs present in slice,
	// we need to exclude test-infra Refs as it probably deliver only tools to test another repo.
	if pjType := message.JobType; *pjType == string(prowapi.PeriodicJob) {
		// If there is only one repo ref, take it.
		if len(message.Refs) == 1 {
			ref = message.Refs[0]
			// There were more refs.
		} else {
			// Find kyma repo ref.
			for _, r := range message.Refs {
				if ref.Repo == "kyma" {
					ref = r
				}
			}
			// There was no Kyma ref in Refs. Take first non test-infra repo ref.
			if ref.Repo == "" {
				for _, r := range message.Refs {
					if r.Repo != TestInfraRepoName {
						ref = r
					}
				}
			}
		}
		// For non periodic prowjobs the first Ref in Refs slice is the one against which prowjob was running.
	} else {
		ref = message.Refs[0]
	}
	// Set data to use for creating github issue.
	issueRequest := &github.IssueRequest{
		// Use failed prowjob name with defined prefix.
		Title: github.String(fmt.Sprintf("Failed prowjob: %s", *message.JobName)),
		Body:  nil,
		// Use common labels.
		Labels:   &[]string{"test-failing", "ci-force-bot"},
		Assignee: nil,
		// Set issue as open.
		State:     github.String("open"),
		Milestone: nil,
		Assignees: nil,
	}
	// Create issue in github.
	issue, ghResponse, err := client.Issues.Create(ctx, ref.Org, ref.Repo, issueRequest)
	// Check if API reply with error.
	if ghResponse != nil {
		err = github.CheckResponse(ghResponse.Response)
		if err != nil {
			return nil, fmt.Errorf("github API call reply with error, error: %w", err)
		}
		// Check if API call failed.
	} else if err != nil {
		return nil, fmt.Errorf("callilg github API failed, error: %w", err)
	}
	return issue, nil
}

// GetGithubIssue is triggered by pubsub message with failing test instance details.
// It will check if received pubsub message data payload has github issue number and create new issue when missing.
// Created github issue number is added to failing test instance document in firestore and pubsub message data payload.
// If github issue is found in a message, it will check if it's status in github is open.
// If closed github issue detected, corresponding failing test instance in firestore will be closed and removed from pubsub message data payload.
// Function then will publish pubsub message back to topic creating new failing test instance in firestore.
// If new github issue was created or existing one is still open, function will publish message for further enriching failing test instance.
func GetGithubIssue(ctx context.Context, m kymapubsub.MessagePayload) error {
	var err error
	// set trace value to use it in logEntry
	var failingTestMessage kymapubsub.FailingTestMessage
	// Create logger to use google cloud functions structured logging
	logger := cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	logger.WithComponent("kyma.prow.cloud-function.GetGithubIssue")
	// Set trace value for log entries to identify messages from one function call.
	logger.GenerateTraceValue(projectID, "GetGithubIssue")

	// Get metadata from context and set eventID label for logging.
	contextMetadata, err := metadata.FromContext(ctx)
	if err != nil {
		if m.MessageId != "" {
			logger.WithLabel("messageId", m.MessageId)
		} else {
			logger.LogError(fmt.Sprintf("failed extract metadata from function call context, error: %s", err.Error()))
		}
	} else {
		logger.WithLabel("messageId", contextMetadata.EventID)
	}

	// Unmarshall pubsub message data payload.
	err = json.Unmarshal(m.Data, &failingTestMessage)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed unmarshal message data field, error: %s", err.Error()))
	}
	// Set label with prowjob name for logging.
	logger.WithLabel("prowjobName", *failingTestMessage.JobName)

	// Set label with execution ID for logging.
	jobID, err := kymapubsub.GetJobId(failingTestMessage.URL)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed get job ID, error: %s", err.Error()))
	}
	logger.WithLabel("jobID", *jobID)
	logger.LogInfo(fmt.Sprintf("found prowjob execution ID: %s", *jobID))

	// Check if failing test message has github issue number.
	if failingTestMessage.GithubIssueNumber != nil {
		// Check if github issue is open. Prevent state when gh issue was closed, but notification bot didn't close it in firestore db.
		open, ghIssue, err := isGithubIssueOpen(ctx, githubClient, failingTestMessage)
		if err != nil {
			logger.LogError(fmt.Sprintf("failed get github issue status, error: %s", err.Error()))
		}
		logger.LogInfo(fmt.Sprintf("github issue number %d has status: %s", ghIssue.GetNumber(), ghIssue.GetState()))
		// When issue is not open
		if !*open {
			// Get firestore document for failing test instance.
			docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, *failingTestMessage.FirestoreDocumentID))
			updates := []firestore.Update{
				{Path: "open", Value: false},
			}
			// Set failing test instance in firestore as closed.
			_, err := docRef.Update(ctx, updates, firestore.Exists)
			// Remove failing test instance firestore document reference from failing test pubsub message.
			failingTestMessage.FirestoreDocumentID = nil
			// Create new github issue.
			ghIssue, err := createGithubIssue(ctx, githubClient, failingTestMessage)
			if err != nil {
				logger.LogError(fmt.Sprintf("failed create github issue, error: %s", err.Error()))
			}
			if ghIssue != nil {
				logger.LogInfo(fmt.Sprintf("github issue created. issue number: %d", ghIssue.GetNumber()))
				// Add created github issue number to failing test pubsub message.
				failingTestMessage.GithubIssueNumber = github.Int64(int64(ghIssue.GetNumber()))
				failingTestMessage.GithubIssueRepo = ghIssue.GetRepository().Name
				failingTestMessage.GithubIssueOrg = ghIssue.GetRepository().GetOwner().Name
				failingTestMessage.GithubIssueUrl = ghIssue.URL
			}
			// Publish message to topic creating new failing test instance in firestore db.
			publlishedMessageID, err := kymapubsub.PublishPubSubMessage(ctx, pubSubClient, failingTestMessage, getFailureInstanceTopic)
			if err != nil {
				logger.LogCritical(fmt.Sprintf("failed publishiing to pubsub, error: %s", err.Error()))
			}
			logger.LogInfo(fmt.Sprintf("published pubsub message to topic %s, id: %s", getFailureInstanceTopic, *publlishedMessageID))
		}
		// Failing tests message has no github issue number.
	} else {
		// Create github issue.
		ghIssue, err := createGithubIssue(ctx, githubClient, failingTestMessage)
		if err != nil {
			logger.LogError(fmt.Sprintf("failed create github issue, error: %s", err.Error()))
		}
		if ghIssue != nil {
			logger.LogInfo(fmt.Sprintf("github issue created. issue number: %d", ghIssue.GetNumber()))
			// Add github issue number to failing test pubsub message.
			failingTestMessage.GithubIssueNumber = github.Int64(int64(*ghIssue.Number))
			failingTestMessage.GithubIssueRepo = ghIssue.GetRepository().Name
			failingTestMessage.GithubIssueOrg = ghIssue.GetRepository().GetOwner().Name
			failingTestMessage.GithubIssueUrl = ghIssue.URL
			docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, *failingTestMessage.FirestoreDocumentID))
			updates := []firestore.Update{
				{Path: "githubIssueNumber", Value: ghIssue.GetNumber()},
				{Path: "githubIssueRepo", Value: ghIssue.GetRepository().GetName()},
				{Path: "githubIssueOrg", Value: ghIssue.GetRepository().GetOwner().GetName()},
				{Path: "githubIssueUrl", Value: ghIssue.GetURL()},
			}
			// Add created github issue details to failing test instance document in firestore.
			_, err := docRef.Update(ctx, updates, firestore.Exists)
			if err != nil {
				logger.LogError(fmt.Sprintf("failed adding github issue number %d, to failing test instance, error: %s", ghIssue.GetNumber(), err.Error()))
				// TODO: need error reporting for such case, without failing whole function
			} else {
				logger.LogInfo(fmt.Sprintf("github issue, number %d, added to failing test instance", ghIssue.GetNumber()))
			}
		} else {
			logger.LogError(fmt.Sprintf("github issue is nil, something went wrong with creating it"))
			// TODO: need error reporting for such case, without failing whole function
		}
	}
	// Publish message to topic further enriching failing test instance.
	commiterPubllishedMessageID, err := kymapubsub.PublishPubSubMessage(ctx, pubSubClient, failingTestMessage, getGithubCommiterTopic)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()))
	}
	logger.LogInfo(fmt.Sprintf("published pubsub message to topic %s, id: %s", getGithubCommiterTopic, *commiterPubllishedMessageID))
	return nil
}
