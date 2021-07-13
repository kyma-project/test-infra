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

type FailingTestMessage struct {
	ProwMessage
	FirestoreDocumentID *string `json:"firestoreDocumentId,omitempty"`
	GithubIssueNumber   *int    `json:"githubIssueNumber,omitempty"`
	SlackThreadID       string  `json:"slackThreadId,omitempty"`
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
	firestoreClient         *firestore.Client
	pubSubClient            *pubsub.Client
	githubClient            *github.Client
	ts                      oauth2.TokenSource
	projectID               string
	githubAccessToken       string
	githubOrg               string
	githubRepo              string
	firestoreCollection     string
	getGithubCommiterTopic  string
	getFailureInstanceTopic string
)

func init() {
	var err error
	projectID = os.Getenv("GCP_PROJECT_ID")
	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	githubOrg = os.Getenv("GITHUB_ORG")
	githubRepo = os.Getenv("GITHUB_REPO")
	firestoreCollection = os.Getenv("FIRESTORE_COLLECTION")
	getGithubCommiterTopic = os.Getenv("GET_GITHUB_COMMITER_TOPIC")
	getFailureInstanceTopic = os.Getenv("GET_FAILURE_INSTANCE_TOPIC")
	ctx := context.Background()
	if getGithubCommiterTopic == "" {
		panic("environment variable GET_GITHUB_COMMITER_TOPIC is empty")
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
	if githubOrg == "" {
		panic("environment variable GITHUB_ORG is empty")
	}
	if githubRepo == "" {
		panic("environment variable GITHUB_REPO is empty")
	}
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating firestore client, error: %s", err.Error()))
	}
	pubSubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating pubsub client, error: %s", err.Error()))
	}
	ts = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient = github.NewClient(tc)
}

func isGithubIssueOpen(ctx context.Context, client *github.Client, message FailingTestMessage, githubOrg, githubRepo, trace, eventID string, jobID *string) (*bool, error) {
	ghIssue, ghResponse, err := client.Issues.Get(ctx, githubOrg, githubRepo, *message.GithubIssueNumber)
	if ghResponse != nil {
		err = github.CheckResponse(ghResponse.Response)
		if err != nil {
			return nil, fmt.Errorf("github API call reply with error, error: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("calling github API failed, error: %w", err)
	}
	log.Println(LogEntry{
		Message:   fmt.Sprintf("github issue number %d has status: %s", message.GithubIssueNumber, *ghIssue.State),
		Severity:  "INFO",
		Trace:     trace,
		Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		Labels:    map[string]string{"messageId": eventID, "jobID": *jobID, "prowjobName": message.JobName},
	})
	b := new(bool)
	if *ghIssue.State == "open" {
		b = github.Bool(true)
	} else {
		b = github.Bool(false)
	}
	return b, nil
}

// TODO: move this function to separate module
// TODO: allign it with getfailurinstance version
func getJobId(jobUrl, eventID, jobName, jobType, trace string) (*string, error) {
	jobURL, err := url.Parse(jobUrl)
	if err != nil {
		return nil, fmt.Errorf("failed parse test URL, error: %w", err)
	}
	jobID := path.Base(jobURL.Path)
	log.Println(LogEntry{
		Message:   fmt.Sprintf("failed %s prowjob detected, prowjob ID: %s", jobType, jobID),
		Severity:  "INFO",
		Trace:     trace,
		Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": jobName},
	})
	return github.String(jobID), nil
}

func createGithubIssue(ctx context.Context, client *github.Client, message FailingTestMessage, githubOrg, githubRepo, trace, eventID string, jobID *string) (*github.Issue, error) {
	issueRequest := &github.IssueRequest{
		Title:     github.String(fmt.Sprintf("Failed prowjob: %s", message.JobName)),
		Body:      nil,
		Labels:    &[]string{"test-failing", "ci-force-bot"},
		Assignee:  nil,
		State:     github.String("open"),
		Milestone: nil,
		Assignees: nil,
	}
	issue, ghResponse, err := client.Issues.Create(ctx, githubOrg, githubRepo, issueRequest)
	if ghResponse != nil {
		err = github.CheckResponse(ghResponse.Response)
		if err != nil {
			return nil, fmt.Errorf("github API call reply with error, error: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("callilg github API failed, error: %w", err)
	}
	log.Println(LogEntry{
		Message:   fmt.Sprintf("github issue created. issue number: %d", issue.Number),
		Severity:  "INFO",
		Trace:     trace,
		Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		Labels:    map[string]string{"messageId": eventID, "jobID": *jobID, "prowjobName": message.JobName},
	})
	return issue, nil
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

func GetGithubIssue(ctx context.Context, m MessagePayload) error {
	var err error
	// set trace value to use it in logEntry
	var trace string
	var failingTestMessage FailingTestMessage
	traceFunctionName := "GetGithubIssue"
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
	jobID, err := getJobId(failingTestMessage.URL, contextMetadata.EventID, failingTestMessage.JobName, failingTestMessage.JobType, trace)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed get job ID, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": contextMetadata.EventID},
		})
		panic(fmt.Sprintf("failed get job ID, error: %s", err.Error()))
	}
	//sprawdz czy message ma gh issue
	if failingTestMessage.GithubIssueNumber != nil {
		//jeśli message ma gh issue sprawdź czy otwarte
		open, err := isGithubIssueOpen(ctx, githubClient, failingTestMessage, githubOrg, githubRepo, trace, contextMetadata.EventID, jobID)
		if err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed get github issue status, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID},
			})
		} else if !*open {
			//stary failure instance w firestore oznacz jako zamknięty
			docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, failingTestMessage.FirestoreDocumentID))
			_, err = docRef.Set(ctx, map[string]bool{"open": false}, firestore.Merge([]string{"open"}))
			//usuń referencje do starego dokumentu w firestore
			failingTestMessage.FirestoreDocumentID = nil
			//jeśli zamknięte to utwórz i dodaj do firestore
			ghIssue, err := createGithubIssue(ctx, githubClient, failingTestMessage, githubOrg, githubRepo, trace, contextMetadata.EventID, jobID)
			if err != nil {
				log.Println(LogEntry{
					Message:   fmt.Sprintf("failed create github issue, error: %s", err.Error()),
					Severity:  "CRITICAL",
					Trace:     trace,
					Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
					Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID, "prowjobName": failingTestMessage.JobName},
				})
			}
			if ghIssue != nil {
				if ghIssue.Number != nil {
					failingTestMessage.GithubIssueNumber = ghIssue.Number
				}
			}
			//opublikuj wiadomość do topicu getfailureinstance
			publlishedMessageID, err := publishPubSubMessage(ctx, pubSubClient, failingTestMessage, getFailureInstanceTopic)
			if err != nil {
				//log error publishing message to pubsub
				log.Println(LogEntry{
					Message:   fmt.Sprintf("failed publishiing to pubsub, error: %s", err.Error()),
					Severity:  "CRITICAL",
					Trace:     trace,
					Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
					Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID, "prowjobName": failingTestMessage.JobName},
				})
				panic(fmt.Sprintf("failed publishiing to pubsub, error: %s", err.Error()))
			}
			// log publishin message to pubsub
			log.Println(LogEntry{
				Message:   fmt.Sprintf("published pubsub message to topic %s, id: %s", getFailureInstanceTopic, *publlishedMessageID),
				Severity:  "INFO",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": failingTestMessage.JobName},
			})
		}
	} else {
		//jeśli message nie ma gh issue to utwórz i dodaj do firestore
		ghIssue, err := createGithubIssue(ctx, githubClient, failingTestMessage, githubOrg, githubRepo, trace, contextMetadata.EventID, jobID)
		if err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed create github issue, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID},
			})
		}
		if ghIssue != nil {
			if ghIssue.Number != nil {
				failingTestMessage.GithubIssueNumber = ghIssue.Number
				docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, *failingTestMessage.FirestoreDocumentID))
				_, err = docRef.Set(ctx, map[string]int{"githubIssueNumber": *ghIssue.Number}, firestore.Merge([]string{"githubIssueNumber"}))
				if err != nil {
					// log error
					log.Println(LogEntry{
						Message:   fmt.Sprintf("failed adding github issue number %d, to failing test instance, error: %s", ghIssue.GetNumber(), err.Error()),
						Severity:  "CRITICAL",
						Trace:     trace,
						Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
						Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID, "prowjobName": failingTestMessage.JobName},
					})
					//TODO: need error reporting for such case, without failing whole function
				} else {
					// log gh issue was added to firestore
					log.Println(LogEntry{
						Message:   fmt.Sprintf("github issue, number %d, added to failing test instance", ghIssue.GetNumber()),
						Severity:  "INFO",
						Trace:     trace,
						Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
						Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": failingTestMessage.JobName},
					})
				}
			} else {
				//log error ghIssue.Number is nil
				log.Println(LogEntry{
					Message:   fmt.Sprintf("github issue number is nil, something went wrong with creating github issue"),
					Severity:  "ERROR",
					Trace:     trace,
					Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
					Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID, "prowjobName": failingTestMessage.JobName},
				})
				//TODO: need error reporting for such case, without failing whole function
			}
		} else {
			//log error getting ghIssue after creating
			log.Println(LogEntry{
				Message:   fmt.Sprintf("github issue is nil, something went wrong with creating it"),
				Severity:  "ERROR",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID, "prowjobName": failingTestMessage.JobName},
			})
			//TODO: need error reporting for such case, without failing whole function
		}
		publlishedMessageID, err := publishPubSubMessage(ctx, pubSubClient, failingTestMessage, getGithubCommiterTopic)
		if err != nil {
			//log error publishing message to pubsub
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()),
				Severity:  "CRITICAL",
				Trace:     trace,
				Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
				Labels:    map[string]string{"jobID": *jobID, "messageId": contextMetadata.EventID, "prowjobName": failingTestMessage.JobName},
			})
			panic(fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()))
		}
		// log publishin message to pubsub
		log.Println(LogEntry{
			Message:   fmt.Sprintf("published pubsub message to topic %s, id: %s", getGithubCommiterTopic, *publlishedMessageID),
			Severity:  "INFO",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": *jobID, "prowjobName": failingTestMessage.JobName},
		})
	}

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
