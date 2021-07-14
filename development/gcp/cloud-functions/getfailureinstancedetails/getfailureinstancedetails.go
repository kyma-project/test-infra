package getfailureinstancedetails

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v36/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	kymapubsub "github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"os"
)

var (
	firestoreClient     *firestore.Client
	pubSubClient        *pubsub.Client
	projectID           string
	getGithubIssueTopic string
)

func init() {
	var err error
	projectID = os.Getenv("GCP_PROJECT_ID")
	getGithubIssueTopic = os.Getenv("GITHUB_ISSUE_TOPIC")
	ctx := context.Background()
	if projectID == "" {
		panic("environment variable GCP_PROJECT_ID is empty, can't setup firebase client")
	}
	if getGithubIssueTopic == "" {
		panic("environment variable GITHUB_ISSUE_TOPIC is empty, can't setup firebase client")
	}
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to create firestore client, error: %s", err.Error()))
	}
	pubSubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to create pubsub client, error: %s", err.Error()))
	}
}

func addFailingTest(ctx context.Context, client *firestore.Client, message kymapubsub.FailingTestMessage, jobID *string) (*firestore.DocumentRef, error) {
	doc, _, err := client.Collection("testFailures").Add(ctx, map[string]interface{}{
		"jobName":           *message.JobName,
		"jobType":           *message.JobType,
		"open":              true,
		"githubIssueNumber": *message.GithubIssueNumber,
		"baseSha":           message.Refs[0]["base_sha"],
		"failures": map[string]interface{}{
			*jobID: map[string]interface{}{
				"url": *message.URL, "gcsPath": *message.GcsPath, "refs": message.Refs,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("colud not add failing test instance to firestore collection, error: %w", err)
	}
	return doc, nil
}

func addTestExecution(ctx context.Context, ref *firestore.DocumentRef, message kymapubsub.FailingTestMessage, jobID *string) error {
	_, err := ref.Set(ctx, map[string]map[string]map[string]interface{}{"failures": {
		*jobID: {
			"url": message.URL, "gcsPath": message.GcsPath, "refs": message.Refs,
		}}}, firestore.Merge([]string{"failures", *jobID}))
	if err != nil {
		return fmt.Errorf("could not add execution data to firestore document, error: %w", err)
	}
	return nil
}

// HelloPubSub consumes a Pub/Sub message.
func GetFailureInstanceDetails(ctx context.Context, m kymapubsub.MessagePayload) error {
	var err error
	// set trace value to use it in logEntry
	var iter *firestore.DocumentIterator
	var failingTestMessage kymapubsub.FailingTestMessage
	logger := cloudfunctions.NewLogger()
	logger = logger.WithComponent("kyma.prow.cloud-function.Getfailureinstancedetails")
	logger = logger.GenerateTraceValue(projectID, "GetFailureInstanceDetails")

	contextMetadata, err := metadata.FromContext(ctx)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed extract metadata from function call context, error: %s", err.Error()))
	}
	logger.WithLabel("messageId", contextMetadata.EventID)

	// Decode
	err = json.Unmarshal(m.Data, &failingTestMessage)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed unmarshal message data field, error: %s", err.Error()))
	}
	logger.WithLabel("prowjobName", *failingTestMessage.JobName)

	if *failingTestMessage.Status == "failure" || *failingTestMessage.Status == "error" {

		if *failingTestMessage.JobType == "periodic" {
			iter = firestoreClient.Collection("testFailures").Where("jobName", "==", *failingTestMessage.JobName).Where("jobType", "==", *failingTestMessage.JobType).Where("open", "==", true).Documents(ctx)
		} else if *failingTestMessage.JobType == "postsubmit" {
			iter = firestoreClient.Collection("testFailures").Where("jobName", "==", *failingTestMessage.JobName).Where("jobType", "==", *failingTestMessage.JobType).Where("open", "==", true).Where("baseSha", "==", failingTestMessage.Refs[0]["base_sha"]).Documents(ctx)
		} else {
			return nil
		}

		jobID, err := kymapubsub.GetJobId(failingTestMessage.URL)
		if err != nil {
			logger.LogCritical(fmt.Sprintf("failed get job ID, error: %s", err.Error()))
		}
		logger.WithLabel("jobID", *jobID)
		logger.LogInfo(fmt.Sprintf("found prowjob execution ID: %s", *jobID))

		failureInstances, err := iter.GetAll()
		if err != nil {
			logger.LogCritical(fmt.Sprintf("failed get failure instances, error: %s", err.Error()))
		}

		if len(failureInstances) == 0 {
			logger.LogInfo("failure instance not found, creating it")

			doc, err := addFailingTest(ctx, firestoreClient, failingTestMessage, jobID)
			if err != nil {
				logger.LogCritical(fmt.Sprintf("could not add failing test, error: %s", err.Error()))
			}
			logger.LogInfo(fmt.Sprintf("failing test created in firestore, document ID: %s", doc.ID))

			failingTestMessage.FirestoreDocumentID = github.String(doc.ID)
		} else if len(failureInstances) == 1 {
			failureInstance := failureInstances[0]
			logger.LogInfo("failure instance exists, adding execution data")

			err = addTestExecution(ctx, failureInstance.Ref, failingTestMessage, jobID)
			if err != nil {
				logger.LogCritical(fmt.Sprintf("failed adding failed test execution data, error: %s", err.Error()))
				// TODO: need error reporting api call
			}

			failingTestMessage.FirestoreDocumentID = github.String(failureInstance.Ref.ID)

			githubIssueNumber, err := failureInstance.DataAt("githubIssueNumber")
			if err != nil {
				logger.LogInfo(fmt.Sprintf("could not get github issue for failing test, error: %s", err.Error()))
			} else {
				failingTestMessage.GithubIssueNumber = github.Int64(githubIssueNumber.(int64))
			}
		} else {
			logger.LogCritical(fmt.Sprintf("more than one failure instance exists for %s prowjob", *failingTestMessage.JobName))
		}

		publlishedMessageID, err := kymapubsub.PublishPubSubMessage(ctx, pubSubClient, failingTestMessage, getGithubIssueTopic)
		if err != nil {
			// log error publishing message to pubsub
			logger.LogCritical(fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()))
		}
		// log publishin message to pubsub
		logger.LogInfo(fmt.Sprintf("published pubsub message to topic %s, id: %s", getGithubIssueTopic, *publlishedMessageID))
	} else {
		logger.LogInfo(fmt.Sprintf("failure not detected, got notification for prowjob %s", *failingTestMessage.JobName))
	}
	return nil
}
