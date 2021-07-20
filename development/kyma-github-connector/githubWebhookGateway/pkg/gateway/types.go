package gateway

//Config containing all program configs
type Config struct {
	GitHubWebhookGatewayName string `envconfig:"GITHUB_WEBHOOK_GATEWAY_NAME"`
	GitHubWebhookSecret      string `envconfig:"GITHUB_WEBHOOK_SECRET"`
	GitHubWebhookUrlPath     string `envconfig:"GITHUB_WEBHOOK_URL_PATH"`
	KymaEventsService        string `envconfig:"EVENTING_SERVICE"` //http://test-gh-connector-app-event-service.kyma-integration:8081/test-gh-connector-app/events
	ListenPort               string `envconfig:"LISTEN_PORT"`
	//EventType         string `envconfig:"EVENT_TYPE"`             // Event type published to Event Publisher Proxy.
	AppName string `envconfig:"APP_NAME"` // PubSub Connector application name as set in Compass.
}
