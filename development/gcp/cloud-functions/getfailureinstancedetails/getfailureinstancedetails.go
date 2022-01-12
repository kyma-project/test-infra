package getfailureinstancedetails

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/pubsub"
	"github.com/google/go-github/v40/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	kymapubsub "github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
)

var (
	firestoreClient      *firestore.Client
	pubSubClient         *pubsub.Client
	projectID            string
	getGithubIssueTopic  string
	firestoreCollection  string
	pjtesterProwjobRegex *regexp.Regexp
)

func init() {
	var err error
	ctx := context.Background()
	// set variables from environment variables
	projectID = os.Getenv("GCP_PROJECT_ID")
	getGithubIssueTopic = os.Getenv("GITHUB_ISSUE_TOPIC")
	firestoreCollection = os.Getenv("FIRESTORE_COLLECTION")
	// check if variables were set with values
	if projectID == "" {
		panic("environment variable GCP_PROJECT_ID is empty, can't setup firebase client")
	}
	if getGithubIssueTopic == "" {
		panic("environment variable GITHUB_ISSUE_TOPIC is empty, can't setup pubsub client")
	}
	if firestoreCollection == "" {
		panic("environment variable FIRESTORE_COLLECTION is empty, can't setup firebase client")
	}
	pjtesterProwjobRegex = regexp.MustCompile(`.+_test_of_prowjob_.+`)
	// create firestore client, it will be reused by multiple function calls
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to create firestore client, error: %s", err.Error()))
	}
	// create pubsub client, it will be reused by multiple function calls
	pubSubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to create pubsub client, error: %s", err.Error()))
	}
}

// addFailingTest creates document for failing prowjob in GCP firestore database.
// Created document represent failing prowjob and holds data about failed runs of prowjob.
func addFailingTest(ctx context.Context, client *firestore.Client, message kymapubsub.FailingTestMessage, jobID *string) (*firestore.DocumentRef, error) {
	// TODO: create struct to represent failing test document
	failingTest := map[string]interface{}{
		// jobName is a failed prowjob name
		"jobName": *message.JobName,
		// jobType is a failed prowjob type
		"jobType": *message.JobType,
		// open indicate if this failure instance is currently active or it's already closed
		"open": true,
		// githubIssueNumber holds Github issue number created for this failure instance
		"githubIssueNumber": nil,
		// baseSha holds sha for which postsubmit prowjob was run.
		"baseSha": message.Refs[0].BaseSHA,
		// failures holds a map with all reported failures of prowjob for which this failure instance was created and active.
		// Entries in a map are a prowjob execution IDs.
		"failures": map[string]interface{}{
			*jobID: map[string]interface{}{
				"url": *message.URL, "gcsPath": *message.GcsPath, "refs": message.Refs,
			},
		},
	}
	if message.GithubIssueNumber != nil {
		// githubIssueNumber holds Github issue number created for this failure instance
		failingTest["githubIssueNumber"] = *message.GithubIssueNumber
	}

	// Add document to firestore collection
	doc, _, err := client.Collection(firestoreCollection).Add(ctx, failingTest)
	if err != nil {
		return nil, fmt.Errorf("colud not add failing test instance to firestore collection, error: %w", err)
	}
	return doc, nil
}

// addTestExecution add failed prowjob execution to already existing and open failure instance in firestore database.
func addTestExecution(ctx context.Context, ref *firestore.DocumentRef, message kymapubsub.FailingTestMessage, jobID *string) error {
	// Add failed test execution to document in firestore db.
	// Failed test execution is added to failures map.
	_, err := ref.Set(ctx, map[string]map[string]map[string]interface{}{"failures": {
		*jobID: {
			"url": message.URL, "gcsPath": message.GcsPath, "refs": message.Refs,
		}}}, firestore.Merge([]string{"failures", *jobID}))
	if err != nil {
		return fmt.Errorf("could not add execution data to firestore document, error: %w", err)
	}
	return nil
}

// GetFailureInstanceDetails is triggered by pubsub message with prowjob details.
// It will detect failed prowjobs and create failing test instance in firestore database if it doesn't exists.
// If firestore database already has open failing test instance for provided prowjob,
// it will add another failed test execution to the instance.
// Function publish pubsub message to topic creating github issues for failing test instances.
func GetFailureInstanceDetails(ctx context.Context, m kymapubsub.MessagePayload) error {
	var err error
	var iter *firestore.DocumentIterator
	var failingTestMessage kymapubsub.FailingTestMessage
	// Create logger to use google cloud functions structured logging
	logger := cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	logger = logger.WithComponent("kyma.prow.cloud-function.Getfailureinstancedetails")
	// Set trace value for log entries to identify messages from one function call.
	logger = logger.GenerateTraceValue(projectID, "GetFailureInstanceDetails")

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

	// Get prowjob execution ID from gcs URL path.
	jobID, err := kymapubsub.GetJobId(failingTestMessage.URL)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed get job ID, error: %s", err.Error()))
	}
	// Set label with execution ID for logging.
	logger.WithLabel("jobID", *jobID)
	logger.LogInfo(fmt.Sprintf("found prowjob execution ID: %s", *jobID))

	if pjtesterProwjobRegex.FindStringIndex(*failingTestMessage.JobName) != nil {
		logger.LogInfo(fmt.Sprintf("prowjob scheduled by pjtester detected, got message for prowjob %s, ignoring", *failingTestMessage.JobName))
		return nil
	}
	// Detect failed prowjobs
	if *failingTestMessage.Status == "failure" || *failingTestMessage.Status == "error" {
		// For periodic prowjob get documents for open failing test instances for matching periodic.
		if *failingTestMessage.JobType == "periodic" {
			iter = firestoreClient.Collection(firestoreCollection).Where(
				"jobName", "==", *failingTestMessage.JobName).Where(
				"jobType", "==", *failingTestMessage.JobType).Where(
				"open", "==", true).Documents(ctx)
			//	For postsubmit prowjob get documents for open failing test for matching postsubmit with the same baseSha. If baseSha is different it's represented by another failing test instance.
		} else if *failingTestMessage.JobType == "postsubmit" {
			iter = firestoreClient.Collection(firestoreCollection).Where(
				"jobName", "==", *failingTestMessage.JobName).Where(
				"jobType", "==", *failingTestMessage.JobType).Where(
				"open", "==", true).Where(
				"baseSha", "==", failingTestMessage.Refs[0].BaseSHA).Documents(ctx)
		} else {
			return nil
		}
		// Get all matched documents retrieved from firestore db.
		failureInstances, err := iter.GetAll()
		if err != nil {
			logger.LogCritical(fmt.Sprintf("failed get failure instances, error: %s", err.Error()))
		}

		// If there was no matching documents create a new test failing instance.
		if len(failureInstances) == 0 {
			logger.LogInfo("failure instance not found, creating it")
			// Add failing test instance to firestore db.
			doc, err := addFailingTest(ctx, firestoreClient, failingTestMessage, jobID)
			if err != nil {
				logger.LogCritical(fmt.Sprintf("could not add failing test, error: %s", err.Error()))
			}
			logger.LogInfo(fmt.Sprintf("failing test created in firestore, document ID: %s", doc.ID))
			// Add created document ID to pubsub message data payload.
			failingTestMessage.FirestoreDocumentID = github.String(doc.ID)
			// If there was one matching document add failing test execution.
		} else if len(failureInstances) == 1 {
			// Get matched document.
			failureInstance := failureInstances[0]
			logger.LogInfo("failure instance exists, adding execution data")
			// Add failing test execution to failing test instance document in firestore db.
			err = addTestExecution(ctx, failureInstance.Ref, failingTestMessage, jobID)
			if err != nil {
				logger.LogCritical(fmt.Sprintf("failed adding failed test execution data, error: %s", err.Error()))
				// TODO: need error reporting api call
			}
			// Add firestore document ID to pubsub message data payload.
			failingTestMessage.FirestoreDocumentID = github.String(failureInstance.Ref.ID)
			// Read github issue number from failing tests instance document.
			githubIssueNumber, err := failureInstance.DataAt("githubIssueNumber")
			if err != nil {
				logger.LogInfo(fmt.Sprintf("could not get github issue for failing test, error: %s", err.Error()))
			}
			if githubIssueNumber != nil {
				// Add github issue number to pubsub message data payload.
				failingTestMessage.GithubIssueNumber = github.Int64(githubIssueNumber.(int64))
			}
		} else {
			logger.LogCritical(fmt.Sprintf("more than one failure instance exists for %s prowjob", *failingTestMessage.JobName))
		}
		// Publish message to pubsub topic.
		publlishedMessageID, err := kymapubsub.PublishPubSubMessage(ctx, pubSubClient, failingTestMessage, getGithubIssueTopic)
		if err != nil {
			logger.LogCritical(fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()))
		}
		logger.LogInfo(fmt.Sprintf("published pubsub message to topic %s, id: %s", getGithubIssueTopic, *publlishedMessageID))
	} else {
		logger.LogInfo(fmt.Sprintf("failure not detected, got message for prowjob %s, ignoring", *failingTestMessage.JobName))
	}
	// Do nothing if prwojob status doesn't mean a failure.
	return nil
}
