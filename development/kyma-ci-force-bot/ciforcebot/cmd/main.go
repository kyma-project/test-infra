package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/v40/github"
	kymaghclient "github.com/kyma-project/test-infra/development/github/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

const (
	errorReportingType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
	githubWebhookIssueClosedV1KymaEventType = "sap.kyma.custom.ciforcebot.issuesevent.closed.v1"
	githubWebhookIssueTransferredV1KymaEventType = "sap.kyma.custom.ciforcebot.issuesevent.transferred.v1"
)

var (
	firestoreClient   *firestore.Client
	loggingClient     *logging.Client
	logger            *logging.Logger
	cloudeventsClient cloudevents.Client
	conf              Config
)

//Config containing all program configs
type Config struct {
	KymaEventsService   string `envconfig:"EVENTING_SERVICE"` //http://test-gh-connector-app-event-service.kyma-integration:8081/test-gh-connector-app/events
	ListenPort          int    `envconfig:"LISTEN_PORT"`
	ProjectID           string `envconfig:"FIRESTORE_GCP_PROJECT_ID"`
	AppName             string `envconfig:"APP_NAME"` // PubSub Connector application name as set in Compass.
	LogName             string `envconfig:"LOG_NAME"` // Google cloud logging log name.
	Component           string `envconfig:"COMPONENT"`
	FirestoreCollection string `envconfig:"FIRESTORE_COLLECTION"`
}

type loggingPayload struct {
	Message   string `json:"message"`
	Operation string `json:"operation,omitempty"`
	Type      string `json:"@type,omitempty"`
}

type GhEvent interface {
	GetIssue() *github.Issue
	GetRepo() *github.Repository
}

func init() {
	var err error
	ctx := context.Background()
	err = envconfig.Init(&conf)
	if err != nil {
		log.Fatalf("failed create config from env variables, got error: %v", err)
	}
	loggingClient, err = logging.NewClient(ctx, conf.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create goggle logging client, got error: %v", err)
	}
	randomInt := rand.Int()
	trace := fmt.Sprintf("projects/%s/traces/%s/%d", conf.ProjectID, conf.Component, randomInt)
	logger := loggingClient.Logger(conf.LogName, logging.CommonLabels(map[string]string{
		"appName":   conf.AppName,
		"component": conf.Component}))
	// create firestore client, it will be reused by multiple function calls
	firestoreClient, err = firestore.NewClient(ctx, conf.ProjectID)
	if err != nil {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Critical,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("failed create firestore client, got error: %v", err),
				Operation: "ci-force-bot initialization on tooling kyma",
				Type:      errorReportingType,
			},
		})
	}
	cloudeventsClient, err = cloudevents.NewClientHTTP(cloudevents.WithPort(conf.ListenPort))
	if err != nil {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Critical,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("failed create firestore client, got error: %v", err),
				Operation: "ci-force-bot initialization on tooling kyma",
				Type:      errorReportingType,
			},
		})
	}
	logger.Flush()
}

func receive(event cloudevents.Event) {
	ctx := context.Background()

	randomInt := rand.Int()
	trace := fmt.Sprintf("projects/%s/traces/%s/%d", conf.ProjectID, conf.Component, randomInt)
	logger := loggingClient.Logger(conf.LogName, logging.CommonLabels(map[string]string{
		"appName":   conf.AppName,
		"component": conf.Component,
		"function":  "closeFailingTestInstance"}))
	defer logger.Flush()
	// do something with event.
	logger.Log(logging.Entry{
		Timestamp: time.Now(),
		Severity:  logging.Info,
		Trace:     trace,
		Payload: loggingPayload{
			Message:   fmt.Sprintf("got event of type %s", event.Type()),
			Operation: "processing received cloudevents event",
		},
	})

	switch event.Type() {
	case githubWebhookIssueClosedV1KymaEventType:
		issueClosedAction(ctx, event,  logger, trace)
	case githubWebhookIssueTransferredV1KymaEventType:
		issueTransferredAction(ctx, event, logger, trace)
	}
}

func decodeGithubEventFromPayload(event cloudevents.Event, issuesEvent GhEvent, logger *logging.Logger, trace string) {
	err := event.DataAs(issuesEvent)

	if err != nil {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Critical,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("loading issue event data failed, got error: %v", err),
				Operation: "processing received cloudevents event",
				Type:      errorReportingType,
			},
		})
	}

	logger.Log(logging.Entry{
		Timestamp: time.Now(),
		Severity:  logging.Info,
		Trace:     trace,
		Payload: loggingPayload{
			Message:   fmt.Sprintf("received %s notification for github issue number: %d", event.Type(), issuesEvent.GetIssue().GetNumber()),
			Operation: "processing received cloudevents event",
		},
	})
}

func getFailingTestInstanceFromFirestore(ctx context.Context, firestoreClient *firestore.Client, issuesEvent GhEvent, logger *logging.Logger, trace string) []*firestore.DocumentSnapshot {

	iter := firestoreClient.Collection(conf.FirestoreCollection).Where(
		"githubIssueNumber", "==", issuesEvent.GetIssue().GetNumber()).Where(
			"githubIssueRepo", "==", issuesEvent.GetRepo().GetName()).Where(
				"githubIssueOrg", "==", issuesEvent.GetRepo().GetOwner().GetName()).Where(
					"open", "==", true).Documents(ctx)
	failureInstances, err := iter.GetAll()

	if err != nil {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Error,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("failed get failure instances from firestore, got error: %v", err),
				Operation: "processing received cloudevents event",
				Type:      errorReportingType,
			},
		})
	}
	return failureInstances
}

func issueTransferredAction(ctx context.Context, event cloudevents.Event, logger *logging.Logger, trace string) {

	issuesEvent := new(kymaghclient.IssueTransferredEvent)
	decodeGithubEventFromPayload(event, issuesEvent, logger, trace)

	failureInstances := getFailingTestInstanceFromFirestore(ctx, firestoreClient, issuesEvent, logger, trace)

	if len(failureInstances) == 1 {
		failureInstance := failureInstances[0]
		updates := []firestore.Update{
			{Path: "githubIssueNumber", Value: issuesEvent.GetChanges().GetNewIssue().GetNumber()},
			{Path: "githubIssueRepo", Value: issuesEvent.GetChanges().GetNewRepository().GetName()},
			{Path: "githubIssueOrg", Value: issuesEvent.GetChanges().GetNewRepository().GetOwner().GetName()},
			{Path: "githubIssueUrl", Value: issuesEvent.GetChanges().GetNewIssue().GetURL()},
		}
		// Set new github issue for failing test instance in firestore.
		_, err := failureInstance.Ref.Update(ctx, updates, firestore.Exists)

		if err != nil {
			logger.Log(logging.Entry{
				Timestamp: time.Now(),
				Severity:  logging.Error,
				Trace:     trace,
				Payload: loggingPayload{
					Message:   fmt.Sprintf("failed replace old github issue %s with new github issue %s in firestore, got error: %v", issuesEvent.GetIssue().GetURL(), issuesEvent.GetChanges().GetNewIssue().GetURL(), err),
					Operation: "processing received cloudevents event",
					Type:      errorReportingType,
				},
			})
		}
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Info,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("updated failing test instance with new github issue in firestore, new github issue: %s", issuesEvent.GetChanges().GetNewIssue().GetURL()),
				Operation: "processing received cloudevents event",
			},
		})
	} else if len(failureInstances) == 0 {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Info,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("could not found open failing test instance for transferred github issue %s", issuesEvent.GetIssue().GetURL()),
				Operation: "processing received cloudevents event",
			},
		})
	} else if len(failureInstances) > 1 {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Error,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("more than one open failing test instance found in firestore for old github issue %s", issuesEvent.GetIssue().GetURL()),
				Operation: "processing received cloudevents event",
				Type:      errorReportingType,
			},
		})
	}
}

func issueClosedAction(ctx context.Context, event cloudevents.Event, logger *logging.Logger, trace string) {

	issuesEvent := new(github.IssuesEvent)
	decodeGithubEventFromPayload(event, issuesEvent, logger, trace)

	failureInstances := getFailingTestInstanceFromFirestore(ctx, firestoreClient, issuesEvent, logger, trace)

	if len(failureInstances) == 1 {
		failureInstance := failureInstances[0]
		updates := []firestore.Update{
			{Path: "open", Value: false},
		}
		// Set failing test instance in firestore as closed.
		_, err := failureInstance.Ref.Update(ctx, updates, firestore.Exists)

		if err != nil {
			logger.Log(logging.Entry{
				Timestamp: time.Now(),
				Severity:  logging.Error,
				Trace:     trace,
				Payload: loggingPayload{
					Message:   fmt.Sprintf("failed set failure instances as closed in firestore for github issue number %d, got error: %v", issuesEvent.GetIssue().GetNumber(), err),
					Operation: "processing received cloudevents event",
					Type:      errorReportingType,
				},
			})
		}
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Info,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("closed failing test instance in firestore for github issue number %d", issuesEvent.GetIssue().GetNumber()),
				Operation: "processing received cloudevents event",
			},
		})
	} else if len(failureInstances) == 0 {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Info,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("could not found open failing test instance for github issue number %d", issuesEvent.GetIssue().GetNumber()),
				Operation: "processing received cloudevents event",
			},
		})
	} else if len(failureInstances) > 1 {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Error,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("more than one open failing test instance found in firestore for github issue number %d", issuesEvent.GetIssue().GetNumber()),
				Operation: "processing received cloudevents event",
				Type:      errorReportingType,
			},
		})
	}
}

func main() {
	randomInt := rand.Int()
	trace := fmt.Sprintf("projects/%s/traces/%s/%d", conf.ProjectID, conf.Component, randomInt)
	logger := loggingClient.Logger(conf.LogName, logging.CommonLabels(map[string]string{
		"appName":   conf.AppName,
		"component": conf.Component}))
	defer logger.Flush()
	logger.Log(logging.Entry{
		Timestamp: time.Now(),
		Severity:  logging.Info,
		Trace:     trace,
		Payload: loggingPayload{
			Message:   fmt.Sprintf("eventing service URL: %s", conf.KymaEventsService),
			Operation: "starting ci-force bot on tooling kyma",
		},
	})
	err := logger.Flush()
	if err != nil {
		log.Errorf("failed send log entry to cloud logging: got error %v", err)
	}

	err = cloudeventsClient.StartReceiver(context.Background(), receive)

	if err != nil {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Critical,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("failed start listening for cloudevents, got error: %v", err),
				Operation: "starting ci-force bot on tooling kyma",
				Type:      errorReportingType,
			},
		})
	}
}
