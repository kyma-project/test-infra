package main

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"google.golang.org/api/option"
)

const (
	credentialsFilePath = "/keys/saCredentials.json"
)

var (
	pubSubClient       *pubsub.Client
	kymeEventClient    cloudevents.Client
	cloudEventsContext context.Context
	conf               Config
)

//Config contain program configs provided bye environment variables.
type Config struct {
	PubSubGatewayName string `envconfig:"PUBSUB_GATEWAY_NAME"`    // Used as eventing cloudevent source.
	KymaEventsService string `envconfig:"EVENTING_SERVICE"`       // URL of Event Publisher Proxy.
	SubscriptionID    string `envconfig:"PUBSUB_SUBSCRIPTION_ID"` // Google PubSub subscription ID to pull messages from.
	EventType         string `envconfig:"EVENT_TYPE"`             // Event type published to Event Publisher Proxy.
	ProjectID         string `envconfig:"PUBSUB_PROJECT_ID"`      // Google PubSub project ID.
	AppName           string `envconfig:"APP_NAME"`               // PubSub Connector application name as set in Compass.
	TargetAppName     string `envconfig:"TARGET_APP_NAME"`        // Kyma eventing target application name where publish events.
}

func main() {
	// Build config from environment variables
	err := envconfig.Init(&conf)
	if err != nil {
		log.Fatalf("failed init config from environment variables %s", err.Error())
	}

	// Create Google PubSub client to pull messages.
	ctx := context.Background()
	pubSubClient, err = pubsub.NewClient(ctx, conf.ProjectID, option.WithCredentialsFile(credentialsFilePath))
	if err != nil {
		log.Fatalf("failed create pubsub client, got error: %s", err.Error())
	}
	defer pubSubClient.Close()

	// Create cloud events client for publishing messages to Kyma.
	kymeEventClient, err = cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed create kyma eventing cloud event client, got error: %s", err.Error())
	}
	cloudEventsContext = cloudevents.ContextWithTarget(context.Background(), conf.KymaEventsService)
	log.Infof("using eventing service URL: %s", conf.KymaEventsService)

	// create subscription client to pull messages from Google PubSub
	sub := pubSubClient.Subscription(conf.SubscriptionID)
	log.Infof("subscribing to pubsub subscription: %s", conf.SubscriptionID)
	ok, err := sub.Exists(ctx)
	if err != nil {
		log.Fatalf("failed to check subscription presence: %v", err.Error())
	}
	log.Infof("subscription %s exists: %t", conf.SubscriptionID, ok)
	// Removing dashes from event type to comply with eventing requirements.
	eventingEventType := strings.Replace(fmt.Sprintf("sap.kyma.custom.%s.%s", conf.TargetAppName, conf.EventType), "-", "", -1)
	log.Infof("using event type : %s", eventingEventType)
	// Create a channel to handle messages from PubSub as they come in.
	cm := make(chan *pubsub.Message)
	defer close(cm)
	// Handle individual messages in a goroutine.
	go func() {
		for msg := range cm {
			log.Infof("received message with ID: %s", msg.ID)
			// Create cloud events event.
			event := cloudevents.NewEvent()
			event.SetSource(conf.PubSubGatewayName)
			event.SetType(eventingEventType)
			event.SetData(cloudevents.ApplicationJSON, msg)
			// Publish event to Kyma eventing.
			if result := kymeEventClient.Send(cloudEventsContext, event); cloudevents.IsUndelivered(result) {
				msg.Nack()
				log.Errorf("failed send event to kyma eventing service: %s", result)
			}
			msg.Ack()
			log.Infof("Message ID %s published to Kyma.", msg.ID)
		}
	}()
	// Pulling messsages from pubsub.
	// Receive blocks until the context is cancelled or an error occurs.
	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		cm <- m
	})
	if err != nil {
		log.Errorf("Finished pulling messages or got an error: %s", err.Error())
	}
}
