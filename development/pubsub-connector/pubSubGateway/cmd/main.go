package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"google.golang.org/api/option"
)

const (
	credentialsFilePath = "/keys/saCredentials.json"
)

var (
	pubSubClient    *pubsub.Client
	kymeEventClient cloudevents.Client
	cloudEventsCTX  context.Context
	conf            Config
)

//Config containing all program configs
type Config struct {
	PubSubGatewayName string `envconfig:"PUBSUB_GATEWAY_NAME"` // used as eventing cloudevent source
	KymaEventsService string `envconfig:"EVENTING_SERVICE"`
	SubscriptionID    string `envconfig:"PUBSUB_SUBSCRIPTION_ID"`
	EventType         string `envconfig:"EVENT_TYPE"`
	ProjectID         string `envconfig:"PUBSUB_PROJECT_ID"`
}

func init() {
	ctx := context.Background()
	var err error
	pubSubClient, err = pubsub.NewClient(ctx, conf.ProjectID, option.WithCredentialsFile(credentialsFilePath))
	if err != nil {
		log.WithFields(log.Fields{"msg": "failed create pubsub client"}).Fatalf("%s", err)
	}
	kymeEventClient, err = cloudevents.NewClientHTTP()
	if err != nil {
		log.WithFields(log.Fields{"msg": "failed create kyma eventing cloud event client"}).Fatalf("%s", err)
	}
	cloudEventsCTX = cloudevents.ContextWithTarget(context.Background(), conf.KymaEventsService)
}

func main() {
	defer pubSubClient.Close()
	err := envconfig.Init(&conf)
	if err != nil {
		log.WithFields(log.Fields{"msg": "failed init config from environment variables"}).Fatalf("%s", err.Error())
	}
	log.WithFields(log.Fields{"msg": "set configuration parameter"}).Infof("Using eventing service URL: %s", conf.KymaEventsService)
	// create subscription to pull messages from
	sub := pubSubClient.Subscription(conf.SubscriptionID)
	ctx := context.Background()
	// Create a channel to handle messages to as they come in.
	cm := make(chan *pubsub.Message)
	defer close(cm)
	// Handle individual messages in a goroutine.
	go func() {
		for msg := range cm {
			log.WithFields(log.Fields{"msg": "received message"}).Infof("message ID:%s", msg.ID)
			event := cloudevents.NewEvent()
			event.SetSource(conf.PubSubGatewayName)
			event.SetType(conf.EventType)
			event.SetData(cloudevents.ApplicationJSON, msg)
			if result := kymeEventClient.Send(cloudEventsCTX, event); cloudevents.IsUndelivered(result) {
				msg.Nack()
				log.WithFields(log.Fields{"msg": "failed send event to kyma eventing service"}).Errorf("%s", result)
			}
			msg.Ack()
		}
	}()

	// Receive blocks until the context is cancelled or an error occurs.
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		cm <- msg
	})
	if err != nil {
		// TODO: support exiting when error and canceling context
		log.WithFields(log.Fields{"msg": "Finished pulling messages or got an error"}).Errorf("%s", err)
	}
}
