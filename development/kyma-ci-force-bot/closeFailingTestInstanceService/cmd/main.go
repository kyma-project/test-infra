package main

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

//Config containing all program configs
type Config struct {
	KymaEventsService string `envconfig:"EVENTING_SERVICE"` //http://test-gh-connector-app-event-service.kyma-integration:8081/test-gh-connector-app/events
	ListenPort        int    `envconfig:"LISTEN_PORT"`
	//EventType         string `envconfig:"EVENT_TYPE"`             // Event type published to Event Publisher Proxy.
	AppName string `envconfig:"APP_NAME"` // PubSub Connector application name as set in Compass.
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
