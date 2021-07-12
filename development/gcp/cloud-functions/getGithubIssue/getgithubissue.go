package getGithubIssue

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v36/github"
	"golang.org/x/oauth2"
	"log"
	"math/rand"
	"os"
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

type FailingTestMessage struct {
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
	firestoreClient   *firestore.Client
	pubSubClient      *pubsub.Client
	githubClient      *github.Client
	ts                oauth2.TokenSource
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
	ts = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv(githubAccessToken)},
	)
	//tc := oauth2.NewClient(ctx, ts)
	//githubClient = github.NewClient(tc)
}

func checkGithubIssueStatus(ctx context.Context, client *github.Client, message ProwMessage, githubOrg, githubRepo, trace, eventID, jobID string, githubIssueNumber interface{}) (*bool, error) {
	issue, response, err := client.Issues.Get(ctx, githubOrg, githubRepo, githubIssueNumber.(int))
	if err != nil {
		log.Println(response)
		log.Println(LogEntry{
			Message:   fmt.Sprintf("could not get github issue number %d, error: %s", githubIssueNumber.(int), err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
		})
		return nil, err
	} else {
		log.Println(response)
		log.Println(LogEntry{
			Message:   fmt.Sprintf("github issue number %d has status: %s", githubIssueNumber.(int), *issue.State),
			Severity:  "INFO",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
		})
		b := new(bool)
		if *issue.State == "open" {
			*b = true
			return b, nil
		}
		return b, nil
	}
}

func GetGithubIssue(ctx context.Context, m MessagePayload) error {
	tc := oauth2.NewClient(ctx, ts)
	githubClient = github.NewClient(tc)
	var err error
	// set trace value to use it in logEntry
	var trace string
	//var iter *firestore.DocumentIterator
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
	//sprawdz czy message ma gh issue
	_, ghResponse, err := githubClient.Issues.Get(ctx, githubOrg, githubRepo, failingTestMessage.GithubIssueNumber)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("github API call failed, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": contextMetadata.EventID, "prowjobName": failingTestMessage.JobName},
		})
	}
	if ghResponse != nil {
		err = github.CheckResponse(ghResponse.Response)
		if err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("github API call reply with error, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "prowjobName": failingTestMessage.JobName},
			})
		}
	}
	rates, _, _ := githubClient.RateLimits(ctx)
	fmt.Println(rates)
	//jeśli message nie ma gh issue to utwórz i dodaj do firestore
	//jeśli message ma gh issue sprawdź czy otwarte
	//jeśli zamknięte to utwórz i dodaj do firestore
	//stary failure instance w firestore oznacz jako zamknięty

	//znajdz commitera - zrób to w osobnej funkcji

	//dodaj do gh issue koemntarz z linkiem do kolejnego wystąpienia błędu, - zrób to w osobnej funkcji
	//link do url
	//nazwa testu
	//czas uruchomienia
	//rodzaj testu
	//base sha
	//pr number
	//commiter
	//jakie logi można dodać w komentarzu. Może oprzeć się na junit lens z prow?
	return nil
}
