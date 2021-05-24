# Overview

This chart lets you install the Github Slack Connector on [Kyma](https://kyma-project.io/).

## Prerequisites

1. GitHub Slack Connector utilizes Kyma features. You need a running Kyma cluster. See the Kyma documentation to learn how to [install](https://kyma-project.io/docs/#installation-installation) it.
2. It uses GitHub Webhook Gateway for receiving webhook events from GitHub. See GitHub Webhook Gateway [installation](../githubWebhookGateway/README.md) to learn how to build it.
3. To validate a source of GitHub Webhook event you need to know GitHub Webhook secret defined for connected repository or organisation.
4. [Helm3](https://helm.sh/docs/intro/install/) installed on your workstation.
5. The Connector publishes messages on a Slack channel. You need a Slack app with the `chat:write` OAuth scope assigned to bot token. See the [Slack documentation](https://api.slack.com/authentication/basics) to learn how to create a new app.

## Installation

1. Clone the `kyma-project/test-infra` repository by running `git clone git@github.com:kyma-project/test-infra.git`.
2. Create [Applications](https://kyma-project.io/docs/components/application-connector/#tutorials-create-a-new-application) resources.
3. Define Slack API [service](https://kyma-project.io/docs/components/application-connector/#tutorials-register-a-service-register-an-api-with-a-specification-url) using the [ghConnectorSlackAPI.json](../../kyma-slack-connector/ghConnectorSlackAPI.json) specification file. Replace `<bot-token>` with your Slack bot token in the `headers` key.
   ```
   "headers": {
                "Authorization": ["Bearer <bot-token>"]
            }
   ```
4. Define a GitHub Events [service](https://kyma-project.io/docs/components/application-connector/#tutorials-register-a-service-register-a-service) using the [ghConnectorEvent.json](ghConnectorEvent.json) specification file.
5.  Run `cd test-infra/development/github-slack-connector`.
6. Create a namespace:
   
   `kubectl create ns gh-connector`
7. Configure Kyma to use [external](https://kyma-project.io/docs/components/serverless/#tutorials-set-an-external-docker-registry) image repository.
8. Now you can install GitHub Slack Connector on your Kyma cluster:
   
   `helm install dekiel04022021 ./githubConnector --set webhookGateway.webhookSecretValue=<yourGithubWebhookSecret> --namespace gh-connector`

## Configuration

Configuration for the GitHub Slack Connector installation can be provided in a [values.yaml](values.yaml) file or `helm install` command. See the comments in the `values.yaml` file for a description of the configuration parameters.

## Development

To check the output from chart templates, execute this command:

`helm install --debug --dry-run test-install --set ghWebhookSecretName=secretName ./githubConnector`
