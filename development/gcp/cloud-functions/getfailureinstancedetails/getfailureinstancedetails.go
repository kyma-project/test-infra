package getfailureinstancedetails

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v36/github"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"
)

// TODO: move types to separate module
// This is the Message payload of pubsub message
type MessagePayload struct {
	Attributes   map[string]string `json:"attributes"`
	Data         []byte            `json:"data"` // This property is base64 encoded
	MessageId    string            `json:"messageId"`
	Message_Id   string            `json:"message_id"`
	PublishTime  string            `json:"publishTime"`
	Publish_time string            `json:"publish_time"`
}

// This is the Data payload of pubsub message payload.
type ProwMessage struct {
	Project *string `json:"project"`
	Topic   *string `json:"topic"`
	RunID   *string `json:"runid"`
	Status  *string `json:"status"`
	URL     *string `json:"url"`
	GcsPath *string `json:"gcs_path"`
	// TODO: define refs type to force using pointers
	Refs    []map[string]interface{} `json:"refs"`
	JobType *string                  `json:"job_type"`
	JobName *string                  `json:"job_name"`
}

type FailingTestMessage struct {
	ProwMessage
	FirestoreDocumentID *string `json:"firestoreDocumentId,omitempty"`
	GithubIssueNumber   *int    `json:"githubIssueNumber,omitempty"`
	SlackThreadID       *string `json:"slackThreadId,omitempty"`
}

// Entry defines a log entry.
type LogEntry struct {
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"`
	// Trace will be the same for one function call, you can use it for filetering in logs
	Trace  string            `json:"logging.googleapis.com/trace,omitempty"`
	Labels map[string]string `json:"logging.googleapis.com/operation,omitempty"`
	// Cloud Log Viewer allows filtering and display of this as `jsonPayload.component`.
	Component string `json:"component,omitempty"`
}

// String renders an entry structure to the JSON format expected by Cloud Logging.
func (e LogEntry) String() string {
	if e.Severity == "" {
		e.Severity = "INFO"
	}
	out, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	return string(out)
}

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

func addFailingTest(ctx context.Context, client *firestore.Client, message FailingTestMessage, jobID *string) (*firestore.DocumentRef, error) {
	doc, _, err := client.Collection("testFailures").Add(ctx, map[string]interface{}{
		"jobName": *message.JobName,
		"jobType": *message.JobType,
		"open":    true,
		"baseSha": message.Refs[0]["base_sha"],
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

func addTestExecution(ctx context.Context, ref *firestore.DocumentRef, message FailingTestMessage, jobID *string) error {
	_, err := ref.Set(ctx, map[string]map[string]map[string]interface{}{"failures": {
		*jobID: {
			"url": message.URL, "gcsPath": message.GcsPath, "refs": message.Refs,
		}}}, firestore.Merge([]string{"failures", *jobID}))
	if err != nil {
		return fmt.Errorf("could not add execution data to firestore document, error: %w", err)
	}
	return nil
}

// TODO: move to separate module
func publishPubSubMessage(ctx context.Context, client *pubsub.Client, message FailingTestMessage, topicName string) (*string, error) {
	bmessage, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling FailingTestMessage to json, error: %w", err)
	}
	topic := client.Topic(topicName)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: bmessage,
	})
	publishedID, err := result.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed publishing to topic %s, error: %w", topicName, err)
	}
	return github.String(publishedID), nil
}

// TODO: move to imported module
func getJobId(message FailingTestMessage, eventID, trace string) (*string, error) {
	jobURL, err := url.Parse(*message.URL)
	if err != nil {
		return nil, fmt.Errorf("failed parse test URL, error: %w", err)
	}
	jobID := path.Base(jobURL.Path)
	log.Println(LogEntry{
		Message:   fmt.Sprintf("failed %s prowjob detected, prowjob ID: %s", *message.JobType, jobID),
		Severity:  "INFO",
		Trace:     trace,
		Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": *message.JobName},
	})
	return github.String(jobID), nil
}

// HelloPubSub consumes a Pub/Sub message.
func Getfailureinstancedetails(ctx context.Context, m MessagePayload) error {
	var err error
	// set trace value to use it in logEntry
	var trace string
	//var prowMessage ProwMessage
	var iter *firestore.DocumentIterator
	var failingTestMessage FailingTestMessage
	traceFunctionName := "Getfailureinstancedetails"
	traceRandomInt := rand.Int()
	trace = fmt.Sprintf("projects/%s/traces/%s/%d", projectID, traceFunctionName, traceRandomInt)
	contextMetadata, err := metadata.FromContext(ctx)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed extract metadata from function call context, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic(fmt.Sprintf("failed extract metadata from function call context, error: %s", err.Error()))
	}
	// Decode
	err = json.Unmarshal(m.Data, &failingTestMessage)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed unmarshal message data field, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": contextMetadata.EventID},
		})
		panic(fmt.Sprintf("failed unmarshal message data field, error: %s", err.Error()))
	}
	//failingTestMessage = FailingTestMessage{
	//	ProwMessage: prowMessage,
	//}
	if *failingTestMessage.Status == "failure" || *failingTestMessage.Status == "error" {
		if *failingTestMessage.JobType == "periodic" {
			iter = firestoreClient.Collection("testFailures").Where("jobName", "==", *failingTestMessage.JobName).Where("jobType", "==", *failingTestMessage.JobType).Where("open", "==", true).Documents(ctx)
		} else if *failingTestMessage.JobType == "postsubmit" {
			iter = firestoreClient.Collection("testFailures").Where("jobName", "==", *failingTestMessage.JobName).Where("jobType", "==", *failingTestMessage.JobType).Where("open", "==", true).Where("baseSha", "==", failingTestMessage.Refs[0]["base_sha"]).Documents(ctx)
		} else {
			return nil
		}
		jobID, err := getJobId(failingTestMessage, contextMetadata.EventID, trace)
		if err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed get job ID, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "prowjobName": *failingTestMessage.JobName},
			})
			panic(fmt.Sprintf("failed get job ID, error: %s", err.Error()))
		}
		failureInstances, err := iter.GetAll()
		if err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed get failure instances, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
			})
			panic(fmt.Sprintf("failed get failure instances, error: %s", err.Error()))
		}
		if len(failureInstances) == 0 {
			log.Println(LogEntry{
				Message:   "failure instance not found, creating it",
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
			})
			doc, err := addFailingTest(ctx, firestoreClient, failingTestMessage, jobID)
			if err != nil {
				log.Println(LogEntry{
					Message:   fmt.Sprintf("could not add failing test, error: %s", err.Error()),
					Severity:  "CRITICAL",
					Trace:     trace,
					Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
					Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
				})
				panic(fmt.Sprintf("could not add failed test, error: %s", err.Error()))
			}
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failing test created in firestore, document ID: %s", doc.ID),
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
			})
			failingTestMessage.FirestoreDocumentID = github.String(doc.ID)
		} else if len(failureInstances) == 1 {
			failureInstance := failureInstances[0]
			log.Println(LogEntry{
				Message:   "failure instance exists, adding execution data",
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
			})
			err = addTestExecution(ctx, failureInstance.Ref, failingTestMessage, jobID)
			if err != nil {
				log.Println(LogEntry{
					Message:   fmt.Sprintf("failed adding failed test execution data, error: %s", err.Error()),
					Severity:  "CRITICAL",
					Trace:     trace,
					Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
					Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
				})
				//TODO: need error reporting api call
			}
			failingTestMessage.FirestoreDocumentID = github.String(failureInstance.Ref.ID)
			githubIssueNumber, err := failureInstance.DataAt("githubIssueNumber")
			if err != nil {
				log.Println(LogEntry{
					Message:   fmt.Sprintf("could not get github issue for failing test, error: %s", err.Error()),
					Severity:  "INFO",
					Trace:     trace,
					Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
					Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
				})
			} else {
				failingTestMessage.GithubIssueNumber = github.Int(githubIssueNumber.(int))
			}
		} else {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("more than one failure instance exists for %s prowjob", *failingTestMessage.JobName),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
			})
			panic(fmt.Sprintf("more than one failure instance exists for %s prowjob", *failingTestMessage.JobName))
		}
		publlishedMessageID, err := publishPubSubMessage(ctx, pubSubClient, failingTestMessage, getGithubIssueTopic)
		if err != nil {
			//log error publishing message to pubsub
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID, "prowjobName": *failingTestMessage.JobName},
			})
			panic(fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()))
		}
		// log publishin message to pubsub
		log.Println(LogEntry{
			Message:   fmt.Sprintf("published pubsub message to topic %s, id: %s", getGithubIssueTopic, *publlishedMessageID),
			Severity:  "INFO",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": *failingTestMessage.JobName},
		})
	}
	log.Println(LogEntry{
		Message:   fmt.Sprintf("failure not detected, got notification for prowjob %s", *failingTestMessage.JobName),
		Trace:     trace,
		Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		Labels:    map[string]string{"messageId": contextMetadata.EventID},
	})
	return nil
}
