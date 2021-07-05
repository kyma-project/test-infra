package getfailureinstancedetails

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"
)

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
	firestoreClient *firestore.Client
	projectID       string
)

func init() {
	var err error
	projectID = os.Getenv("GCP_PROJECT_ID")
	ctx := context.Background()
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

// HelloPubSub consumes a Pub/Sub message.
func Getfailureinstancedetails(ctx context.Context, m MessagePayload) error {
	var err error
	// set trace value to use it in logEntry
	var trace string
	var prowMessage ProwMessage
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
	}
	// Decode
	err = json.Unmarshal(m.Data, &prowMessage)
	if err != nil {
		fmt.Println("error:", err)
	}
	if prowMessage.JobType == "periodic" || prowMessage.JobType == "postsubmit" {
		if prowMessage.Status == "failure" || prowMessage.Status == "error" {
			jobURL, err := url.Parse(prowMessage.URL)
			if err != nil {
				log.Println(LogEntry{
					Message:   "failed parse test URL",
					Severity:  "CRITICAL",
					Trace:     trace,
					Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
					Labels:    map[string]string{"messageId": contextMetadata.EventID},
				})
			}
			jobID := path.Base(jobURL.Path)
			iter := firestoreClient.Collection("testFailures").Where("jobName", "==", prowMessage.JobName).Where("jobType", "==", prowMessage.JobType).Where("open", "==", true).Documents(ctx)
			failureInstances, err := iter.GetAll()
			if err != nil {
				log.Println(LogEntry{
					Message:   fmt.Sprintf("failed get failure instances, error: %s", err.Error()),
					Severity:  "CRITICAL",
					Trace:     trace,
					Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
					Labels:    map[string]string{"messageId": contextMetadata.EventID},
				})
			}
			if len(failureInstances) == 1 {
				// TODO: design how to extract and store commitIDs
				_, _, err = firestoreClient.Collection("testFailures").Add(ctx, map[string]interface{}{
					"jobName": prowMessage.JobName,
					"jobType": prowMessage.JobType,
					"open":    true,
					"failures": map[string]interface{}{
						jobID: map[string]interface{}{
							"url": prowMessage.URL, "gcsPath": prowMessage.GcsPath, "refs": prowMessage.Refs,
						},
					},
				})
				if err != nil {
					log.Println(LogEntry{
						Message:   "failed add failure instance to the collection",
						Severity:  "CRITICAL",
						Trace:     trace,
						Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
						Labels:    map[string]string{"messageId": contextMetadata.EventID},
					})
				}
			}
		}
	}
	return nil
}
