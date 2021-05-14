# Overview

This chart lets you install the PubSub Connector on [Kyma](https://kyma-project.io/).

## Prerequisites

1. PubSub Connector utilizes Kyma features. You need a running Kyma cluster. See the Kyma documentation to learn how to [install](https://kyma-project.io/docs/#installation-installation) it.
2. It uses PubSub Gateway for pulling messages from PubSub subscription. See PubSub Gateway [installation](../pubSubGateway/README.md) to learn how to build it.
3. [Helm3](https://helm.sh/docs/intro/install/) installed on your workstation.
4. The Connector publishes messages on a Slack channel. You need a Slack app with the `chat:write` OAuth scope assigned to bot token. See the [Slack documentation](https://api.slack.com/authentication/basics) to learn how to create a new app.

## Installation

1. Clone the `kyma-project/test-infra` repository by running `git clone git@github.com:kyma-project/test-infra.git`.
2. Create [Applications](https://kyma-project.io/docs/components/application-connector/#tutorials-create-a-new-application) resources.
3. Define Slack API [service](https://kyma-project.io/docs/components/application-connector/#tutorials-register-a-service-register-an-api-with-a-specification-url).
4. Run `cd test-infra/development/pubsub-connector`.
5. Create a namespace: `kubectl create ns gh-connector`
6. Configure Kyma to use [external](https://kyma-project.io/docs/components/serverless/#tutorials-set-an-external-docker-registry) image repository.
7. Now you can install PubSub Connector on your Kyma cluster:
   
   `helm install dekiel04022021 ./pubsubConnector --set slackConnector.botToken="xoxb-xxxx-xxxx-xxxx-xxxx" --namespace pubsub-connector`

## Configuration

Configuration for the PubSub Connector installation can be provided in a [values.yaml](values.yaml) file or `helm install` command. See the comments in the `values.yaml` file for a description of the configuration parameters.

## Development

To check the output from chart templates, execute this command:

`helm install --debug --dry-run test-install --set slackConnector.botToken="xoxb-xxxx-xxxx-xxxx-xxxx" ./githubConnector`
