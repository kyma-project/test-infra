package main

import (
	"net/http"

	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/events"
	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/github"
	"github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway/pkg/handlers"

	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

//Config containing all program configs
type Config struct {
	GitHubWebhookGatewayName string `envconfig:"GITHUB_WEBHOOK_GATEWAY_NAME"`
	GitHubWebhookSecret      string `envconfig:"GITHUB_WEBHOOK_SECRET"`
	KymaEventsService        string `envconfig:"EVENTING_SERVICE"` //http://test-gh-connector-app-event-service.kyma-integration:8081/test-gh-connector-app/events
	ListenPort               string `envconfig:"LISTEN_PORT"`
	EventsPort               string `envconfig:"EVENTING_PORT"`
}

func main() {
	var conf Config
	err := envconfig.Init(&conf)
	if err != nil {
		log.Fatal("Env error: ", err.Error())
	}
	log.Infof("Eventing service URL: %s", conf.KymaEventsService)
	log.Infof("Port: %s", conf.EventsPort)

	kyma := events.NewSender(&http.Client{}, events.NewValidator(), conf.KymaEventsService)
	webhook := handlers.NewWebHookHandler(
		github.NewReceivingEventsWrapper(conf.GitHubWebhookSecret),
		kyma,
	)

	http.HandleFunc("/webhook", webhook.HandleWebhook)
	log.Info(http.ListenAndServe(":"+conf.ListenPort, nil))

	log.Info("Happy GitHub-Connecting!")

}
