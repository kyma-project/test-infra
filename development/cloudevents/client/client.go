package client

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/v42/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/logging"
	console "github.com/kyma-project/test-infra/development/logging"
	"golang.org/x/net/context"
)

type Config struct {
	ListenPort int `envconfig:"LISTEN_PORT"`
}

type Client struct {
	cloudevents.Client
	listenPort    int
	mux           func(cloudevents.Event)
	eventHandlers map[string]func(*Client, cloudevents.Event)
	Logger        logging.LoggerInterface
}

type Option func(client *Client) error

func NewClient(options ...Option) (*Client, error) {
	cc := &Client{
		Client:        nil,
		listenPort:    8080,
		mux:           nil,
		eventHandlers: nil,
		Logger:        console.NewLogger(),
	}

	for _, opt := range options {
		err := opt(cc)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	client, err := cloudevents.NewClientHTTP(cloudevents.WithPort(cc.listenPort))
	if err != nil {
		return nil, fmt.Errorf("failed craeting cloudevents HTTP client: %w", err)
	}
	cc.Client = client
	return cc, nil
}

func WithListenPort(listenPort int) Option {
	return func(cc *Client) error {
		cc.listenPort = listenPort
		return nil
	}
}

func WithDefaultMux() Option {
	return func(cc *Client) error {
		cc.mux = cc.defaultMux
		return nil
	}
}

func FromConfig(config Config) func(*Client) error {
	return func(cc *Client) error {
		cc.listenPort = config.ListenPort
		return nil
	}
}

func WithEventHandler(eventType string, eventHandler func(*Client, cloudevents.Event)) Option {
	return func(cc *Client) error {
		if _, ok := cc.eventHandlers[eventType]; !ok {
			cc.eventHandlers[eventType] = eventHandler
		} else {
			return fmt.Errorf("event handler for event type %s already registered", eventType)
		}
		return nil
	}
}

func WithLogger(logger logging.LoggerInterface) Option {
	return func(cc *Client) error {
		cc.Logger = logger
		return nil
	}
}

func (cc *Client) Run(ctx context.Context) error {
	err := cc.StartReceiver(ctx, cc.mux)
	if err != nil {
		return fmt.Errorf("failed start receiver: %w", err)
	}
	return nil
}

func (cc *Client) defaultMux(event cloudevents.Event) {
	cc.Logger.Infof("received %s event", event.Type())
	if eventHandler, ok := cc.eventHandlers[event.Type()]; ok {
		eventHandler(cc, event)
	} else {
		cc.Logger.Infof("skipping unsupported event %s", event.Type())
	}
}

func (cc *Client) DecodeGithubEventFromCloudEventPayload(event cloudevents.Event) (github.Event, error) {
	var ghEvent github.Event
	err := event.DataAs(ghEvent)
	return ghEvent, err
}
