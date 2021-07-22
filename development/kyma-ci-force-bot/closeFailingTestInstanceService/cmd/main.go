package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/gateway"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

var (
	firestoreClient *firestore.Client
	conf            Config
)

//Config containing all program configs
type Config struct {
	KymaEventsService string `envconfig:"EVENTING_SERVICE"` //http://test-gh-connector-app-event-service.kyma-integration:8081/test-gh-connector-app/events
	ListenPort        int    `envconfig:"LISTEN_PORT"`
	ProjectID         string `envconfig:"FIRESTORE_GCP_PROJECT_ID"`
	//EventType         string `envconfig:"EVENT_TYPE"`             // Event type published to Event Publisher Proxy.
	AppName string `envconfig:"APP_NAME"` // PubSub Connector application name as set in Compass.
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
}

func receive(event cloudevents.Event) {
	// do something with event.
	issueEvent := new(github.IssuesEvent)
	err := event.DataAs(issueEvent)
	if err != nil {
		log.Errorf("load issue event data failed")
	}
	log.Infof("%d", issueEvent.Issue.GetNumber())
}

func main() {
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
