module github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.6.1
	github.com/google/go-github/v40 v40.0.0
	github.com/kyma-project/test-infra/development/github-slack-connector/githubWebhookGateway v0.0.0-20210528091155-95eb95378149
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)
