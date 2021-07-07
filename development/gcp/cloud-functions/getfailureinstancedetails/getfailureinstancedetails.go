package getfailureinstancedetails

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"
)

//TODO: move types to separate module
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
	Project string                   `json:"project"`
	Topic   string                   `json:"topic"`
	RunID   string                   `json:"runid"`
	Status  string                   `json:"status"`
	URL     string                   `json:"url"`
	GcsPath string                   `json:"gcs_path"`
	Refs    []map[string]interface{} `json:"refs"`
	JobType string                   `json:"job_type"`
	JobName string                   `json:"job_name"`
}

type FailingTest struct {
	ProwMessage
	FirestoreDocumentID string `json:"firestoreDocumentId,omitempty"`
	GithubIssueNumber   int    `json:"githubIssueNumber,omitempty"`
	SlackThreadID       string `json:"slackThreadId,omitempty"`
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
	ctx := context.Background()
	if projectID == "" {
		log.Println(LogEntry{
			Message:   "environment variable GCP_PROJECT_ID is empty, can't setup firebase client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GCP_PROJECT_ID is empty, can't setup firebase client")
	}
	if getGithubIssueTopic == "" {
		log.Println(LogEntry{
			Message:   "environment variable GITHUB_ISSUE_TOPIC is empty, can't setup firebase client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GITHUB_ISSUE_TOPIC is empty, can't setup firebase client")
	}
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed create firestore client, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic(fmt.Sprintf("Failed to create firestore client, error: %s", err.Error()))
	}
	pubSubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed create pubsub client, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic(fmt.Sprintf("Failed to create pubsub client, error: %s", err.Error()))
	}
}

func addFailingTest(ctx context.Context, client *firestore.Client, message ProwMessage, jobID, trace, eventID string) error {
	_, _, err := client.Collection("testFailures").Add(ctx, map[string]interface{}{
		"jobName": message.JobName,
		"jobType": message.JobType,
		"open":    true,
		"baseSha": message.Refs[0]["base_sha"],
		"failures": map[string]interface{}{
			jobID: map[string]interface{}{
				"url": message.URL, "gcsPath": message.GcsPath, "refs": message.Refs,
			},
		},
	})
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("could not add failing test, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
		})
		return fmt.Errorf("could not add failing test, error: %w", err)
	}
	return nil
}

func addTestExecution(ctx context.Context, ref *firestore.DocumentRef, message ProwMessage, jobID, trace, eventID string) error {
	_, err := ref.Set(ctx, map[string]map[string]map[string]interface{}{"failures": {
		jobID: {
			"url": message.URL, "gcsPath": message.GcsPath, "refs": message.Refs,
		}}}, firestore.Merge([]string{"failures", jobID}))
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("could not add execution data to failing test, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
		})
		return fmt.Errorf("could not add execution data to failing test, error: %w", err)
	}
	return nil
}

func publishPubSubMessage(ctx context.Context, client *pubsub.Client, message FailingTest, trace, eventID, jobID string) error {
	bmessage, err := json.Marshal(message)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed marshal failing test message, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
		})
		return fmt.Errorf("failed marshal failing test message, error: %w", err)
	}
	topic := client.Topic(getGithubIssueTopic)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: bmessage,
	})
	publishedID, err := result.Get(ctx)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed publish failing test message, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
		})
		return fmt.Errorf("failed publish failing test message, error: %w", err)
	}
	log.Println(LogEntry{
		Message:   fmt.Sprintf("published failing test message, id: %s", publishedID),
		Severity:  "INFO",
		Trace:     trace,
		Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
	})
	return nil
}

// HelloPubSub consumes a Pub/Sub message.
func Getfailureinstancedetails(ctx context.Context, m MessagePayload) error {
	var err error
	// set trace value to use it in logEntry
	var trace string
	var prowMessage ProwMessage
	var iter *firestore.DocumentIterator
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
	err = json.Unmarshal(m.Data, &prowMessage)
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
	if prowMessage.Status == "failure" || prowMessage.Status == "error" {
		if prowMessage.JobType == "periodic" {
			iter = firestoreClient.Collection("testFailures").Where("jobName", "==", prowMessage.JobName).Where("jobType", "==", prowMessage.JobType).Where("open", "==", true).Documents(ctx)
		} else if prowMessage.JobType == "postsubmit" {
			iter = firestoreClient.Collection("testFailures").Where("jobName", "==", prowMessage.JobName).Where("jobType", "==", prowMessage.JobType).Where("open", "==", true).Where("baseSha", "==", prowMessage.Refs[0]["base_sha"]).Documents(ctx)
		} else {
			return nil
		}
		jobURL, err := url.Parse(prowMessage.URL)
		if err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed parse test URL, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID},
			})
			panic(fmt.Sprintf("failed parse test URL, error: %s", err.Error()))
		}
		jobID := path.Base(jobURL.Path)
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed %s prowjob detected, prowjob ID: %s", prowMessage.JobType, jobID),
			Severity:  "INFO",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
		})
		failureInstances, err := iter.GetAll()
		if err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed get failure instances, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
			})
			panic(fmt.Sprintf("failed get failure instances, error: %s", err.Error()))
		}
		if len(failureInstances) == 0 {
			log.Println(LogEntry{
				Message:   "failure instance not found, creating",
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
			})
			err = addFailingTest(ctx, firestoreClient, prowMessage, jobID, trace, contextMetadata.EventID)
			if err != nil {
				panic(err.Error())
			}
		} else if len(failureInstances) == 1 {
			log.Println(LogEntry{
				Message:   "failure instance exists, adding execution data",
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
			})
			err = addTestExecution(ctx, failureInstances[0].Ref, prowMessage, jobID, trace, contextMetadata.EventID)
			if err != nil {
				panic(err.Error())
			}
		} else {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("more than one failure instance exists for %s prowjob", prowMessage.JobName),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
			})
			panic(fmt.Sprintf("more than one failure instance exists for %s prowjob", prowMessage.JobName))
		}
		githubIssueNumber, err := failureInstance.DataAt("githubIssueNumber")
		if err != nil {
			log.Println(LogEntry{
				Message:   "github issue for failing test doesn't exists",
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
			})
		}
		failingTest := FailingTest{
			ProwMessage:         prowMessage,
			FirestoreDocumentID: "",
			GithubIssueNumber:   0,
			SlackThreadID:       "",
		}
		publishPubSubMessage(ctx, pubSubClient, failing, trace, contextMetadata.EventID, jobID)
	}
	return nil
}
