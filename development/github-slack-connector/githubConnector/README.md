# Overview

This chart lets' you install Github Slack Connector on [Kyma](https://kyma-project.io/).

## Prerequisites

1. Github Slack Connector utilises Kyma features. You need a running Kyma cluster. See Kyma documentation to learn how to [install](https://kyma-project.io/docs/#installation-installation) it.
2. It uses Github Webhook Gateway for receiving webhook events from Github. See Github Webhook Gateway [installation](../githubWebhookGateway/README.md) to learn how to build it.
3. To validate a source of github webhook event you need to know github webhook secret defined for connected repository or organisation.
4. helm3 installed on your workstation.
5. Connector publish messages on slack channel. You need a slack app with `chat:write` OAuth scope assigned to bot token. See slack [docs](https://api.slack.com/authentication/basics) to learn how to create new app.

## Installation

1. Clone kyma-project/test-infra repository `git clone git@github.com:kyma-project/test-infra.git`
2. Create [Applications](https://kyma-project.io/docs/components/application-connector/#tutorials-create-a-new-application) resources.
3. Define Slack API [service](https://kyma-project.io/docs/components/application-connector/#tutorials-register-a-service-register-an-api-with-a-specification-url) using [ghConnectorSlackAPI.json](ghConnectorSlackAPI.json) specification file. Replace `<bot-token>` with your Slack bot token in **headers** key.
   ```
   "headers": {
                "Authorization": ["Bearer <bot-token>"]
            }
   ```
4. Define Github Events [service](https://kyma-project.io/docs/components/application-connector/#tutorials-register-a-service-register-a-service) using [ghConnectorEvent.json](ghConnectorEvent.json) specification file.
5. `cd test-infra/development/github-slack-connector`
6. Create namespace.
   
   `kubectl create ns gh-connector`
7. Configure Kyma to [use external](https://kyma-project.io/docs/components/serverless/#tutorials-set-an-external-docker-registry) image repository.
8. Now you can install Github Slack Connector on Kyma cluster.
   
   `helm install dekiel04022021 ./githubConnector --set webhookGateway.webhookSecretValue=<yourGithubWebhookSecret> --namespace gh-connector`

## Configuration

Configuration for Github Slack Connector installation can be provided in [values.yaml](values.yaml) file or in helm install command. See comments in values.yaml file for parameters description.

## Development

To check output from chart templates execute following command.

`helm install --debug --dry-run test-install --set ghWebhookSecretName=secretName ./githubConnector`
