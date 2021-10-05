package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

const (
	errorReportingType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
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

	issueEvent := new(github.IssuesEvent)
	err := event.DataAs(issueEvent)

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
			Message:   fmt.Sprintf("received issue closed notification for issue number: %d", issueEvent.Issue.GetNumber()),
			Operation: "processing received cloudevents event",
		},
	})

	iter := firestoreClient.Collection("testFailures").Where("open", "==", true).Where("githubIssueNumber", "==", issueEvent.Issue.GetNumber()).Documents(ctx)
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

	if len(failureInstances) == 1 {
		failureInstance := failureInstances[0]
		// Set failing test instance in firestore as closed.
		_, err := failureInstance.Ref.Set(ctx, map[string]bool{"open": false}, firestore.Merge([]string{"open"}))

		if err != nil {
			logger.Log(logging.Entry{
				Timestamp: time.Now(),
				Severity:  logging.Error,
				Trace:     trace,
				Payload: loggingPayload{
					Message:   fmt.Sprintf("failed set failure instances as closed in firestore for github issue number %d, got error: %v", issueEvent.Issue.GetNumber(), err),
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
				Message:   fmt.Sprintf("closed failing test instance in firestore for github issue number %d", issueEvent.Issue.GetNumber()),
				Operation: "processing received cloudevents event",
			},
		})
	} else if len(failureInstances) == 0 {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Info,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("could not found open failing test instance for github issue number %d", issueEvent.Issue.GetNumber()),
				Operation: "processing received cloudevents event",
			},
		})
	} else if len(failureInstances) > 1 {
		logger.Log(logging.Entry{
			Timestamp: time.Now(),
			Severity:  logging.Error,
			Trace:     trace,
			Payload: loggingPayload{
				Message:   fmt.Sprintf("more than one open failing test instance found in firestore for github issue number %d", issueEvent.Issue.GetNumber()),
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
