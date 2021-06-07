# GitHub Webhook Gateway

## Overview

The GitHub Webhook Gateway is written in Golang. It listens for incoming GitHub Webhook events. It validates a Webhook HMAC secret, converts it to a CloudEvents compliant event and forwards it to the Kyma Event Service. It is meant to run within Kyma as a docker container.

## Prerequisites

Github Webhook Gateway was created to run within the Kyma Runtime, but it will work with any [CloudEvents](https://github.com/cloudevents/spec/blob/v1.0/spec.md) compliant receiver.

## Installation

To install github-webhook-gateway binary, follow these steps:

1. `git clone git@github.com:kyma-project/test-infra.git`
2. `cd test-infra/development/github-slack-connector/githubWebhookGateway`
3. `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ghwebhookgateway ./cmd/main.go`

Use a [Dockerfile](Dockerfile) to build a Docker image:

1. `git clone git@github.com:kyma-project/test-infra.git`
2. `cd test-infra/development/github-slack-connector/githubWebhookGateway`
3. `docker build .`

## Usage

Github Webhook Gateway uses environment variables to read the configuration.

| Environment variable name | Required | Description |
|----------------|----------|-------------|
| **GITHUB_WEBHOOK_GATEWAY_NAME** | Yes | Github Webhook Gateway instance name. It will be used as a cloud event sourceID |
| **GITHUB_WEBHOOK_SECRET** | Yes | Github Webhook event secret. Used to validate source of a event |
| **EVENTING_SERVICE** | Yes | URL of Kyma Event Service or any CloudEvents compliant receiver |
| **LISTEN_PORT** | Yes | Port number on which Github Webhook Gateway will listen for incoming webhook events |
| **EVENTING_PORT** | Yes | Port number of Kyma Event Service or any CloudEvents compliant receiver |

Github Webhook Gateway expects to get webhook events on the `/webhook` HTTP path.

## Development

To run tests, use this command:

`go test ./...`
