# Overview

The Function oomFound is executed by Kyma [Serverless](https://kyma-project.io/docs/components/serverless/) when registered **oomevent.found** event occur. Function is written in Python.

## Prerequisites

You need Kyma Runtime to run this function.

## Installation

Push the [handler.py](handler.py) and [requirements.txt](requirements.txt) files to your GitHub repository and [point Serverless](https://kyma-project.io/docs/components/serverless/#tutorials-create-a-function-from-git-repository-sources) to download its code for building a Function image.

## Configuration

Function is using [ServiceBindingUsage](https://kyma-project.io/docs/components/serverless/#tutorials-bind-a-service-instance-to-a-function) to get Slack API URL. To point function to the environment variable provided by ServiceBindingUsage set **SLACK_API_ID** environment variable by providing it's value in *slackConnector.apiId* key in [values.yaml](../pubsubConnector/values.yaml). Slack api ID will be used to build target [variable name](https://github.com/kyma-project/test-infra/blob/5f6e98dd0cf692156a0d1caf9b3df5c27d39368d/development/pubsub-connector/oomFoundFunction/handler.py#L15).

Function use **SLACK_BOT_TOKEN** environment variable to authorise sending messages to slack channels. To use valid token, set it on chart installation with helm command argument `--set slackConnector.botToken="xoxb-xxx-xxxx-xxxx-xxxx-xxx"` or set it in [values.yaml](../pubsubConnector/values.yaml).

To send your messages to chosen channel, provide its name in **NOTIFICATION_SLACK_CHANNEL** environment variable. This variable is set by *function.oomevent.found.notificationSlackChannel* key in [values.yaml](../pubsubConnector/values.yaml).
