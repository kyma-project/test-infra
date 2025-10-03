package client

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/v75/github"
	"github.com/kyma-project/test-infra/pkg/logging"
	"golang.org/x/net/context"
)

// Config holds configuration for Client.
// It can be passed to the client constructor with FromConfig client configuration option.
type Config struct {
	ListenPort int `envconfig:"LISTEN_PORT"`
}

// Client is a cloudevents client implementation wrapper for sending and receiving cloudevents.
// It wraps cloudevents v2 client.
type Client struct {
	cloudevents.Client
	listenPort    int
	mux           func(cloudevents.Event)
	eventHandlers map[string]func(*Client, cloudevents.Event)
	Logger        logging.LoggerInterface
}

// Option is a client constructor configuration option for passing configuration to the client constructor.
type Option func(*Client) error

// NewClient is a constructor function for cloudevents client.
// A constructed client can be configured by providing Options.
// Default constructed logger is a console logger.
// Default listen port is 8080.
func NewClient(options ...Option) (*Client, error) {
	cc := &Client{
		Client:        nil,
		listenPort:    8080,
		mux:           nil,
		eventHandlers: nil,
		Logger:        logging.NewLogger(),
	}

	for _, opt := range options {
		err := opt(cc)
		if err != nil {
			return nil, fmt.Errorf("failed to set option: %w", err)
		}
	}

	client, err := cloudevents.NewClientHTTP(cloudevents.WithPort(cc.listenPort))
	if err != nil {
		return nil, fmt.Errorf("failed craeting cloudevents HTTP client: %w", err)
	}
	cc.Client = client
	return cc, nil
}

// WithListenPort is a client constructor configuration option specifying port number to listen on for a cloudevents.
// If not provided default listen port is 8080.
func WithListenPort(listenPort int) Option {
	return func(cc *Client) error {
		cc.listenPort = listenPort
		return nil
	}
}

// WithDefaultMux is a client constructor configuration option enabling client default mux.
func WithDefaultMux() Option {
	return func(cc *Client) error {
		cc.mux = cc.defaultMux
		return nil
	}
}

// FromConfig is a client constructor configuration option passing client Config struct.
// Data from struct will be used to construct client.
func FromConfig(config Config) Option {
	return func(cc *Client) error {
		cc.listenPort = config.ListenPort
		return nil
	}
}

// WithEventHandler is a client constructor configuration option adding event handler for event type.
// Client use this handler to process every received event of assigned eventType.
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

// WithLogger is a client constructor configuration option passing logger instance.
// If not provided default constructed logger is a console logger.
func WithLogger(logger logging.LoggerInterface) Option {
	return func(cc *Client) error {
		cc.Logger = logger
		return nil
	}
}

// Run start a cloudevents receiver and listen for incoming events.
func (cc *Client) Run(ctx context.Context) error {
	err := cc.StartReceiver(ctx, cc.mux)
	if err != nil {
		return fmt.Errorf("failed start receiver: %w", err)
	}
	return nil
}

// defaultMux is a default muxer. It routes received cloudevents to assigned handlers.
// Assigned handlers are taken from eventHandlers map.
func (cc *Client) defaultMux(event cloudevents.Event) {
	cc.Logger.Infof("received %s event", event.Type())
	if eventHandler, ok := cc.eventHandlers[event.Type()]; ok {
		eventHandler(cc, event)
	} else {
		cc.Logger.Infof("skipping unsupported event %s", event.Type())
	}
}

// DecodeGithubEvent retrieve and return a respective GitHub event from cloudevents event.
func (cc *Client) DecodeGithubEvent(event cloudevents.Event) (github.Event, error) {
	var ghEvent github.Event
	err := event.DataAs(ghEvent)
	return ghEvent, err
}

func (cc *Client) RegisterEvent(eventType string, handler func(*Client, cloudevents.Event)) error {
	if _, ok := cc.eventHandlers[eventType]; !ok {
		cc.eventHandlers[eventType] = handler
	} else {
		return fmt.Errorf("event handler for event type %s already registered", eventType)
	}
	return nil
}
