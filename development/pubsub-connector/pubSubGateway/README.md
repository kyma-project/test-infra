# PubSub Gateway

## Overview

The PubSub Gateway is written in Golang. It pulls messages from PubSub, converts them into a CloudEvents compliant event and forwards it to the Kyma Event Service.

## Prerequisites

PubSub Gateway is meant to run in a Pod within the Kyma Runtime.

## Installation

Use a [Dockerfile](Dockerfile) to build a Docker image:

1. `git clone git@github.com:kyma-project/test-infra.git`
2. `cd test-infra/development/pubsub-connector/pubSubGateway`
3. `docker build .`
4. `docker push`

## Usage

PubSub Gateway uses environment variables to read the configuration.

| Environment variable name | Required | Description |
|----------------|----------|-------------|
| **APP_NAME** | Yes | PubSub connector application as set in Compass |
| **PUBSUB_GATEWAY_NAME** | Yes | PubSub Gateway instance name. It will be used as a cloud event sourceID. |
| **PUBSUB_SUBSCRIPTION_ID** | Yes | PubSub subscription ID to pull messages from. |
| **PUBSUB_PROJECT_ID** | Yes | Project ID where PubSub subscription to pull messages from exists. |
| **EVENTING_SERVICE** | Yes | URL of Kyma Eventing Proxy Service. |
| **EVENT_TYPE** | Yes | CloudEvents event type. |
