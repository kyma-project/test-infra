# Overview

This chart lets you install the PubSub Connector on [Kyma](https://kyma-project.io/). The PubSub Connector pulls messages from a PubSub subscription through the PubSub Gateway. It translates the messages to Cloud Events and publishes them to the Kyma Event Publisher Proxy. These events are consumed by the Kyma Serverless [*oomFoundFunction*](../oomFoundFunction), which extracts data from the event and sends a notification to a Slack channel. The Function utilizes the Slack Connector Application APIs to send notifications.

## Prerequisites

1. PubSub Connector utilizes Kyma features. You need a running Kyma cluster. See the Kyma documentation to learn how to [install](https://kyma-project.io/docs/#installation-installation) it.
2. The PubSub Connector uses the PubSub Gateway to pull messages from a PubSub subscription. See the [PubSub Gateway installation](../pubSubGateway/README.md) document to learn how to build it.
3. You need to have [Helm3](https://helm.sh/docs/intro/install/) installed on your workstation.
4. The Connector publishes messages on a Slack channel. You need a Slack app with the `chat:write` OAuth scope assigned to a bot token. See the [Slack documentation](https://api.slack.com/authentication/basics) to learn how to create a new app.

## Installation

1. Clone the `kyma-project/test-infra` repository by running `git clone git@github.com:kyma-project/test-infra.git`.
2. Create [Application](https://kyma-project.io/docs/components/application-connector/#tutorials-create-a-new-application) resources.
3. Define a Slack [API service](https://kyma-project.io/docs/components/application-connector/#tutorials-register-a-service-register-an-api-with-a-specification-url).
4. Run `cd test-infra/development/pubsub-connector`.
5. Create a Namespace by running `kubectl create ns gh-connector`
6. Configure Kyma to use an [external image repository](https://kyma-project.io/docs/components/serverless/#tutorials-set-an-external-docker-registry).
7. Install the PubSub Connector on your Kyma cluster:
   
   `helm install dekiel04022021 ./pubsubConnector --set slackConnector.botToken="xoxb-xxxx-xxxx-xxxx-xxxx" --namespace pubsub-connector`

## Configuration

Configuration for the PubSub Connector installation can be provided in a [values.yaml](values.yaml) file or `helm install` command. See the comments in the `values.yaml` file for a description of the configuration parameters.

## Development

To check the output from chart templates, execute this command:

`helm install --debug --dry-run test-install --set slackConnector.botToken="xoxb-xxxx-xxxx-xxxx-xxxx" ./githubConnector`
