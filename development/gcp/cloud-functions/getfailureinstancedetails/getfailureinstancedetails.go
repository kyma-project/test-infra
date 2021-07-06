package getfailureinstancedetails

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v36/github"
	"golang.org/x/oauth2"
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
	firestoreClient   *firestore.Client
	githubClient      *github.Client
	projectID         string
	githubAccessToken string
	githubOrg         string
	githubRepo        string
)

func init() {
	var err error
	projectID = os.Getenv("GCP_PROJECT_ID")
	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	githubOrg = os.Getenv("GITHUB_ORG")
	githubRepo = os.Getenv("GITHUB_REPO")
	ctx := context.Background()
	if projectID == "" {
		log.Println(LogEntry{
			Message:   "environment variable GCP_PROJECT_ID is empty, can't setup firebase client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GCP_PROJECT_ID is empty, can't setup firebase client")
	}
	if githubAccessToken == "" {
		log.Println(LogEntry{
			Message:   "environment variable GITHUB_ACCESS_TOKEN is empty, can't setup github client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GITHUB_ACCESS_TOKEN is empty, can't setup github client")
	}
	if githubOrg == "" {
		log.Println(LogEntry{
			Message:   "environment variable GITHUB_ORG is empty, can't setup github client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GITHUB_ACCESS_TOKEN is empty, can't setup github client")
	}
	if githubRepo == "" {
		log.Println(LogEntry{
			Message:   "environment variable GITHUB_REPO is empty, can't setup github client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GITHUB_ACCESS_TOKEN is empty, can't setup github client")
	}
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed create firestore client, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic(fmt.Sprintf("Failed to create client, error: %s", err.Error()))
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv(githubAccessToken)},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubClient = github.NewClient(tc)
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
	}
	return err
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
	}
	return err
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
		}
		if len(failureInstances) == 0 {
			log.Println(LogEntry{
				Message:   "failure instance not found, creating",
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
			})
			_ = addFailingTest(ctx, firestoreClient, prowMessage, jobID, trace, contextMetadata.EventID)
		} else if len(failureInstances) == 1 {
			//TODO: check if instance is closed in githuub
			githubIssueNumber, err := failureInstances[0].DataAt("githubIssueNumber")
			if err != nil {
				log.Println("gh isse not found")
			} else {
				issue, _, err := githubClient.Issues.Get(ctx, githubOrg, githubRepo, githubIssueNumber.(int))
				if err != nil {
					log.Println(err.Error())
				} else {
					println(issue.State)
				}
			}
			log.Println(LogEntry{
				Message:   "failure instance exists, adding execution data",
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
			})
			_ = addTestExecution(ctx, failureInstances[0].Ref, prowMessage, jobID, trace, contextMetadata.EventID)
		} else {
			for _, failureInstance := range failureInstances {
				githubIssueNumber, err := failureInstance.DataAt("githubIssueNumber")
				if err != nil {
					log.Println("gh issue not found")
				} else {
					issue, _, err := githubClient.Issues.Get(ctx, githubOrg, githubRepo, githubIssueNumber.(int))
					if err != nil {
						log.Println(err.Error())
					} else {
						println(issue.State)
					}
				}
			}
			//TODO: check if instance is closed in githuub This should not happen
			log.Println(LogEntry{
				Message:   fmt.Sprintf("more than one failure instance exist for periodic %s prowjob", prowMessage.JobName),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
			})
		}
	}
	return nil
}
