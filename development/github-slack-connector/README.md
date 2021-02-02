# The GitHub and Slack Connectors for Kyma

---

[![Go Report Card](https://goreportcard.com/badge/github.com/dekiel/github-slack-connectors)](https://goreportcard.com/report/github.com/dekiel/github-slack-connectors)

---

## Overview

This document describes the Connectors for GitHub and Slack to use in the [Kyma](https://github.com/kyma-project/kyma) environment. They allow utilizing applications' functions inside the Kyma ecosystem by communicating with the corresponding APIs. Use them to trigger lambda functions on Events incoming from third-party applications and react to them.

## Prerequisites

* **Kyma**
The Connectors are configured to work inside the Kyma ecosystem, so you must install them locally or on a cluster. See the [Installation guides](https://kyma-project.io/docs/root/kyma#installation-installation) for details.

## Usage

You can [install](#quick-start) an example scenario, which labels issues on GitHub that may be offensive and sends notifications to Slack about it. However, considering the fact that the Connectors provide a way to communicate with external applications, there are many possible use cases. Using the Connectors is as simple as deploying a new lambda function in Kyma. Check the corresponding [serverless documentation](https://kyma-project.io/docs/components/serverless) to find out more.

This diagram shows the interaction of the components in the described scenario:

![Software architecture image](docs/assets/flowdiagram.svg)

## Quick start

You can install the Connectors and start using them in just a few steps. Follow the instructions to install the Connectors and run the described scenario.

1. Add Add-Ons configuration to Kyma. Run:

    ``` shell
    cat <<EOF | kubectl apply -f -
    apiVersion: addons.kyma-project.io/v1alpha1
    kind: ClusterAddonsConfiguration
    metadata:
      name: addons-slack-github-connectors
      finalizers:
      - addons.kyma-project.io
    spec:
      repositories:
        - url: github.com/dekiel/github-slack-connectors//addons/index.yaml
        - url: github.com/dekiel/github-slack-connectors//addons/index-scenario.yaml
    EOF
    ```

2. Connect to the Kyma Console (UI). Go to a Namespace of your choice, then to **Catalog** in the **Service Management** section. Add the Slack Connector, the GitHub Connector, and the Azure Service Broker. Follow the instructions available in these Add-Ons.
3. After provisioning, add the GitHub Issue Sentiment Analysis Scenario.

    >**NOTE:** Keep in mind that all resources created in the previous step must be ready before you proceed. Check their status in **Instances** in the **Service Management** section.

4. Create a new issue on the GitHub repository specified during the GitHub Connector installation to check if everything is configured correctly. After you create the issue, its sentiment is checked and if it is negative, you get a notification on Slack, and the issue is tagged with the `Caution/offensive` label.

## Installation

Install the Connectors locally or on a cluster. For installation details, see the corresponding guides:

* [The GitHub Connector installation](/docs/github-connector/installation.md)
* [The Slack Connector installation](/docs/slack-connector/installation.md)

## Development

1. Fork the repository in GitHub.
2. Clone the fork to your `$GOPATH` workspace. Use this command to create the folder structure and clone the repository under the correct location:

    ``` shell
    git clone git@github.com:{GitHubUsername}/github-slack-connectors.git $GOPATH/src/github.com/dekiel/github-slack-connectors
    ```

    Follow the steps described in the [`git-workflow.md`](https://github.com/kyma-project/community/blob/master/contributing/03-git-workflow.md) document to configure your fork.
3. Install dependencies in the main project directory. For example, for the GitHub Connector run:

    ``` shell
    cd github-connector
    dep ensure -vendor-only
    ```
