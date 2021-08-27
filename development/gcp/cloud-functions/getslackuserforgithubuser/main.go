package getslackuserforgithubuser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	gcpfirestore "cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	"github.com/kyma-project/test-infra/development/gcp/pkg/firestore"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/types"
)

var (
	pubSubClient        *pubsub.Client
	githubClient        *client.SapToolsClient
	firestoreClient     *firestore.Client
	githubAccessToken   string
	projectID           string
	githubOrg           string
	githubRepo          string
	notifyCommiterTopic string
	firestoreCollection string
)

func init() {
	var err error
	ctx := context.Background()
	// set variables from environment variables
	projectID = os.Getenv("GCP_PROJECT_ID")
	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	githubOrg = os.Getenv("GITHUB_ORG")
	githubRepo = os.Getenv("GITHUB_REPO")
	notifyCommiterTopic = os.Getenv("NOTIFY_COMMITER_TOPIC")
	firestoreCollection = os.Getenv("FIRESTORE_COLLECTION")
	// check if variables were set with values
	if notifyCommiterTopic == "" {
		panic("environment variable NOTIFY_COMMITER_TOPIC is empty")
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
	if firestoreCollection == "" {
		panic("environment variable FIRESTORE_COLLECTION is empty, can't setup firebase client")
	}
	// create github client
	githubClient, err = client.NewSapToolsClient(ctx, githubAccessToken)
	if err != nil {
		panic(fmt.Sprintf("Failed creating github client, error: %s", err.Error()))
	}
	// create pubsub client
	pubSubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating pubsub client, error: %s", err.Error()))
	}
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("failed creating firestore client, error: %s", err.Error()))
	}
}

// GetSlackUserForGithubUser finds slack username for all github usernames from FailingTestMessage.GithubCommitersLogins.
// It search for slack usernames in users-map.yaml file in github. Results are stored in FailingTestMessage.CommitersSlackLogin.
// It publish message to pubsub NotifyCommiterTopic and update failing prowjob instance in firestore.
func GetSlackUserForGithubUser(ctx context.Context, m pubsub.MessagePayload) error {
	var err error
	var wg sync.WaitGroup
	out := make(chan string)
	done := make(chan int)
	// set trace value to use it in logEntry
	var failingTestMessage pubsub.FailingTestMessage

	// TODO: Move setting up logging for cloudfunction to separate method in logging package
	// Create logger to use google cloud functions structured logging
	logger := cloudfunctions.NewLogger()
	// Set component for log entries to identify all messages for this function.
	// TODO: pass function name as constant or variable.
	logger.WithComponent("kyma.prow.cloud-function.GetSlackUserForGithubUser")
	// Set trace value for log entries to identify messages from one function call.
	logger.GenerateTraceValue(projectID, "GetSlackUserForGithubUser")

	// Get metadata from context and set eventID label for logging.
	contextMetadata, err := metadata.FromContext(ctx)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed extract metadata from function call context, error: %s", err.Error()))
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
	jobID, err := pubsub.GetJobId(failingTestMessage.URL)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed get job ID, error: %s", err.Error()))
	} else {
		logger.WithLabel("jobID", *jobID)
		logger.LogInfo(fmt.Sprintf("found prowjob execution ID: %s", *jobID))
	}

	// Check if FailingTestMessage.CommitersSlackLogins is empty. If not empty, nothing to do.
	// Only pass message to pubsub NotifyCommiterTopic.
	if failingTestMessage.CommitersSlackLogins == nil || len(failingTestMessage.CommitersSlackLogins) == 0 {
		// Get content of users-map.yaml.
		usersMap, err := githubClient.GetUsersMap(ctx)
		if err != nil {
			logger.LogCritical(fmt.Sprintf("failed get users-map.yaml file from sap tools github, got error: %v", err))
		}

		// Increase waiting group with number github commiters to check.
		wg.Add(len(failingTestMessage.GithubCommitersLogins))
		for _, commiter := range failingTestMessage.GithubCommitersLogins {
			// Run goroutine to find entry in usersMap for one commiter. One goroutine for each commiter.
			go func(commiter string, usersMap []types.User, logger *cloudfunctions.LogEntry, out chan<- string, done chan<- int) {
				// Signal close channel when done.
				defer func() { done <- 1 }()
				// Search usersMap.
				for _, user := range usersMap {
					if user.ComGithubUsername == commiter {
						logger.LogInfo(fmt.Sprintf("user %s is present in users map", commiter))
						// Send slack username to output channel.
						// TODO: consider creating an User type to store matching github and slack user in one object.
						out <- user.ComEnterpriseSlackUsername
						return
					}
				}
				logger.LogError(fmt.Sprintf("user %s is not present in users map, please add user to users-map.yaml", commiter))
			}(commiter, usersMap, logger, out, done)
		}

		// Routine receiving results from search go routines. It saves results in FailingTestMessage.CommiterSlackLogins.
		go func(wg *sync.WaitGroup, message *pubsub.FailingTestMessage, logger *cloudfunctions.LogEntry, out <-chan string, done <-chan int) {
			for {
				select {
				case slackUser := <- out:
					logger.LogInfo(fmt.Sprintf("adding slack user %s to CommitersSlacklogins", slackUser))
					// Save slack username in FailingTestMessage.CommiterSlackLogins
					message.CommitersSlackLogins = append(message.CommitersSlackLogins, slackUser)
				case <-done:
					// Decrease waiting group count.
					wg.Done()
				}
			}
		}(&wg, &failingTestMessage, logger, out, done)
		// Wait until all search goroutines finish.
		wg.Wait()
		// Check if we found all authors.
		if len(failingTestMessage.GithubCommitersLogins) == len(failingTestMessage.CommitersSlackLogins) {
			logger.LogInfo("all authors present in users map")
		}
		// Update failing prowjob instance in firestore with slack usernames.
		var ref *gcpfirestore.DocumentRef
		// If FailingTestMessage hold firestore ID, use it.
		if failingTestMessage.FirestoreDocumentID != nil {
			// Create firestore document ref.
			ref = firestoreClient.Doc(fmt.Sprintf("testFailures/%s", *failingTestMessage.FirestoreDocumentID))
			err = firestoreClient.StoreSlackUsernames(ctx, failingTestMessage.CommitersSlackLogins, ref)
			if err != nil {
				logger.LogError(fmt.Sprintf("failed store commiters slack usernames in firestore, got error: %s", err.Error()))
			}
		} else {
			// TODO: confirm this will be recorded as error in google cloud error reporting
			logger.LogError(fmt.Sprintf("failingTestMessage.FirestoreDocumentID is empty, can not store commiters slack usernames in firestore"))
		}
	}

	// Publish message to pubsub.
	publlishedMessageID, err := pubSubClient.PublishMessage(ctx, failingTestMessage, notifyCommiterTopic)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()))
	} else {
		logger.LogInfo(fmt.Sprintf("published pubsub message to topic %s, id: %s", notifyCommiterTopic, *publlishedMessageID))
	}
	return nil
}
