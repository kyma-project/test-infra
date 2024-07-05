package firestore

import (
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/pkg/gcp/pubsub"

	"cloud.google.com/go/firestore"
)

const (
	// This is a default google project ID for pubsub workloads.
	//PubSubProjectID = "sap-kyma-prow"
	// TODO: change to sap-kyma-prow for production usage
	PubSubProjectID = "sap-kyma-neighbors-dev"
)

// NewClient create kyma implementation of pubsub Client.
// It wraps google pubsub client.
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	firestoreClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create google firestore client, got error: %w", err)
	}
	return &Client{Client: firestoreClient}, nil
}

func (c *Client) GetFailingProwjobInstanceDetails(ctx context.Context, message pubsub.FailingTestMessage, firestoreCollectionName string) (*firestore.DocumentSnapshot, error) {
	var iter *firestore.DocumentIterator
	// For periodic prowjob get documents for open failing test instances with matching periodic prowjob name.
	if *message.JobType == "periodic" {
		// TODO: rename collection to prowjobFailures
		iter = c.Collection(firestoreCollectionName).Where("jobName", "==", *message.JobName).Where("jobType", "==", *message.JobType).Where("open", "==", true).Documents(ctx)
		//	For postsubmit prowjob get documents for open failing test with matching baseSha. If baseSha is different it's represented by another failing prowjob instance.
	} else if *message.JobType == "postsubmit" {
		iter = c.Collection(firestoreCollectionName).Where("jobName", "==", *message.JobName).Where("jobType", "==", *message.JobType).Where("open", "==", true).Where("baseSha", "==", message.Refs[0].BaseSHA).Documents(ctx)
	} else {
		return nil, fmt.Errorf("got message for presubmit prowjob, storing failing prowjob instance details are not supported for this type of prowjob")
	}
	// Get all matched documents fetched from firestore db.
	failureInstances, err := iter.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed get failing prowjob instance documents, got error: %w", err)
	}

	if len(failureInstances) == 0 {
		return nil, fmt.Errorf("failing prowjob instance not found")
	} else if len(failureInstances) == 1 {
		// Get matched document.
		failureInstance := failureInstances[0]
		return failureInstance, nil
	}
	return nil, fmt.Errorf("more than one failure instance exists")
}

func (c *Client) StoreSlackUsernames(ctx context.Context, slackUserNames []string, ref *firestore.DocumentRef) error {
	// Add failed test execution to document in firestore db.
	// Failed test execution is added to failures map.
	update := []firestore.Update{
		{
			Path:  "commitersSlackLogins",
			Value: slackUserNames,
		},
	}
	_, err := ref.Update(ctx, update)
	if err != nil {
		return fmt.Errorf("could not add execution data to firestore document, error: %w", err)
	}
	return nil
}
