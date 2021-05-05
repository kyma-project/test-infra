package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"google.golang.org/api/option"
)

const (
	credentialsFilePath      = "/keys/saCredentials.json"
	debugCredentialsFilePath = "/Users/i319037/Downloads/saCredentials.json"
)

var (
	pubSubClient       *pubsub.Client
	kymeEventClient    cloudevents.Client
	cloudEventsContext context.Context
	conf               Config
	err                error
)

//Config containing all program configs
type Config struct {
	PubSubGatewayName string `envconfig:"PUBSUB_GATEWAY_NAME"` // used as eventing cloudevent source
	KymaEventsService string `envconfig:"EVENTING_SERVICE"`
	SubscriptionID    string `envconfig:"PUBSUB_SUBSCRIPTION_ID"`
	EventType         string `envconfig:"EVENT_TYPE"`
	ProjectID         string `envconfig:"PUBSUB_PROJECT_ID"`
	AppName           string `envconfig:"APP_NAME"`
}

func main() {
	// Build config from environment variables
	err := envconfig.Init(&conf)
	if err != nil {
		log.Fatalf("failed init config from environment variables %s", err.Error())
	}
	// make a
	ctx := context.Background()
	pubSubClient, err = pubsub.NewClient(ctx, conf.ProjectID, option.WithCredentialsFile(credentialsFilePath))
	//pubSubClient, err = pubsub.NewClient(ctx, conf.ProjectID, option.WithCredentialsFile(debugCredentialsFilePath))
	if err != nil {
		log.WithFields(log.Fields{"msg": "failed create pubsub client"}).Fatalf("%s", err)
	}
	defer pubSubClient.Close()
	kymeEventClient, err = cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed create kyma eventing cloud event client, got error: %s", err)
	}
	cloudEventsContext = cloudevents.ContextWithTarget(context.Background(), conf.KymaEventsService)
	log.Infof("using eventing service URL: %s", conf.KymaEventsService)
	// create subscription to pull messages from
	sub := pubSubClient.Subscription(conf.SubscriptionID)
	log.Infof("subscribing to pubsub subscription: %s", conf.SubscriptionID)
	ok, err := sub.Exists(ctx)
	if err != nil {
		log.Fatalf("failed to check subscription presence: %v", err)
	}
	log.Infof("subscription exists: %t", ok)
	log.Infof("subscribing to %s", conf.SubscriptionID)
	eventingEventType := fmt.Sprintf("sap.kyma.custom.m-%s.%s", conf.AppName, conf.EventType)
	log.Infof("using event type : %s", eventingEventType)
	// Create a channel to handle messages to as they come in.
	cm := make(chan *pubsub.Message)
	defer close(cm)
	// Handle individual messages in a goroutine.
	go func() {
		for msg := range cm {
			log.WithFields(log.Fields{"msg": "received message"}).Infof("message ID:%s", msg.ID)
			event := cloudevents.NewEvent()
			event.SetSource(conf.PubSubGatewayName)
			event.SetType(eventingEventType)
			event.SetData(cloudevents.ApplicationJSON, msg)
			if result := kymeEventClient.Send(cloudEventsContext, event); cloudevents.IsUndelivered(result) {
				msg.Nack()
				log.WithFields(log.Fields{"msg": "failed send event to kyma eventing service"}).Errorf("%s", result)
			}
			msg.Ack()
		}
	}()

	// Receive blocks until the context is cancelled or an error occurs.
	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		cm <- m
	})
	if err != nil {
		// TODO: support exiting when error and canceling context
		log.WithFields(log.Fields{"msg": "Finished pulling messages or got an error"}).Errorf("%s", err)
	}
}
