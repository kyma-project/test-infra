package main

import (
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/events"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/gateway"
	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/github"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

func main() {
	var conf gateway.Config
	err := envconfig.Init(&conf)
	if err != nil {
		log.Fatal("failed create config from env variables: ", err.Error())
	}
	log.Infof("eventing service URL: %s", conf.KymaEventsService)

	client, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed create cloudevents client: %s", err.Error())
	}

	// create sender which is responsible for sending cloudevents messages to kyma
	kyma := events.NewSender(client, events.NewValidator(), conf.KymaEventsService, conf.AppName)
	webhook := gateway.NewWebHookHandler(
		github.NewReceivingEventsWrapper(conf.GitHubWebhookSecret),
		kyma,
	)

	http.HandleFunc(conf.GitHubWebhookURLPath, webhook.HandleWebhook)
	log.Info(http.ListenAndServe(":"+conf.ListenPort, nil))

	log.Info("Happy GitHub-Connecting!")

}
