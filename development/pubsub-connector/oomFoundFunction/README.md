# Overview

The Function oomFound is executed by Kyma [Serverless](https://kyma-project.io/docs/components/serverless/) when registered **oomevent.found** event occur.

## Prerequisites

You need Kyma Runtime to run this function.

## Installation

Push the [handler.py](handler.py) and [requirements.txt](requirements.txt) files to your GitHub repository and [point Serverless](https://kyma-project.io/docs/components/serverless/#tutorials-create-a-function-from-git-repository-sources) to download its code for building a Function image.

## Configuration

Function is using [ServiceBindingUsage](https://kyma-project.io/docs/components/serverless/#tutorials-bind-a-service-instance-to-a-function) to get Slack API URL. To point function to the environment variable provided by ServiceBindingUsage set slackConnector.apiId in [values.yaml](../pubsubConnector/values.yaml).

Function use SLACK_BOT_TOKEN environment variable to authorise sending messages to slack channels. To set valid token, use slackConnector.botToken in [values.yaml](../pubsubConnector/values.yaml).
