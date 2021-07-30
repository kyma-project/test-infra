package main

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"time"
)

var (
	firestoreClient *firestore.Client
	loggingClient   *logging.Client
	logger          *logging.Logger
	conf            Config
)

//Config containing all program configs
type Config struct {
	KymaEventsService string `envconfig:"EVENTING_SERVICE"` //http://test-gh-connector-app-event-service.kyma-integration:8081/test-gh-connector-app/events
	ListenPort        int    `envconfig:"LISTEN_PORT"`
	ProjectID         string `envconfig:"FIRESTORE_GCP_PROJECT_ID"`
	//EventType         string `envconfig:"EVENT_TYPE"`             // Event type published to Event Publisher Proxy.
	AppName string `envconfig:"APP_NAME"` // PubSub Connector application name as set in Compass.
	LogName string `envconfig:"LOG_NAME"` // Google cloud logging log name.
}

type loggingPayload struct {
	Message   string
	Component string
	Operation string
}

func init() {
	var err error
	ctx := context.Background()
	err = envconfig.Init(&conf)
	if err != nil {
		log.Fatal("failed create config from env variables: ", err.Error())
	}
	// create firestore client, it will be reused by multiple function calls
	firestoreClient, err = firestore.NewClient(ctx, conf.ProjectID)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed creating firestore client, error: %s", err.Error()))
	}
	loggingClient, err = logging.NewClient(ctx, conf.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create goggle logging client: %v", err)
	}
	logger = loggingClient.Logger(conf.LogName, logging.CommonLabels(map[string]string{
		"appName":  conf.AppName,
		"function": "closeFailingTestInstance"}))
}

func receive(event cloudevents.Event) {
	defer logger.Flush()
	// do something with event.
	ctx := context.Background()
	log.Infof("got event of type %s", event.Type())
	issueEvent := new(github.IssuesEvent)
	err := event.DataAs(issueEvent)
	if err != nil {
		log.Errorf("load issue event data failed")
	}
	log.Infof("%d", issueEvent.Issue.GetNumber())
	iter := firestoreClient.Collection("testFailures").Where("open", "==", true).Where("githubIssueNumber", "==", issueEvent.Issue.GetNumber()).Documents(ctx)
	failureInstances, err := iter.GetAll()
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed get failure instances, error: %s", err.Error()))
	}
	if len(failureInstances) == 1 {
		failureInstance := failureInstances[0]
		// Set failing test instance in firestore as closed.
		_, err = failureInstance.Ref.Set(ctx, map[string]bool{"open": false}, firestore.Merge([]string{"open"}))
		// TODO: add comment on github issue about closing test instance with respective number.
		// TODO: add logging to stackdriver.
	} else if len(failureInstances) == 0 {
		log.Infof("could not found open failing test instance for github issue number %d", issueEvent.Issue.GetNumber())
	} else if len(failureInstances) > 1 {
		// TODO: Report failure to stackdriver.
		log.Fatalf("to many open failing test instance found in firestore")
	}
}

func main() {
	defer logger.Flush()
	logger.Log(logging.Entry{
		Timestamp: time.Now(),
		Severity:  logging.Info,
		Payload: loggingPayload{
			Message:   "Test log message",
			Component: "kyma.tooling.ci-force-bot",
			Operation: "closeFailingTestInstance",
		},
		Trace: "testTrace",
	})
	var conf Config
	err := envconfig.Init(&conf)
	if err != nil {
		log.Fatal("failed create config from env variables: ", err.Error())
	}
	log.Infof("eventing service URL: %s", conf.KymaEventsService)

	client, err := cloudevents.NewClientHTTP(cloudevents.WithPort(conf.ListenPort))
	if err != nil {
		log.Fatalf("failed create cloudevents client: %s", err.Error())
	}
	err = client.StartReceiver(context.Background(), receive)
	if err != nil {
		log.Errorf("failed listen for cloudevents, error: %s", err.Error())
	}
}
