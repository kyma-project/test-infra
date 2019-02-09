# Prow Addons Controller Manager

## Overview

The Prow Addons Controller Manager embeds all custom controller extension for Prow infrastructure. This project is bootstrapped by Kubebuilder. Check the official [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder/tree/v1.0.7) documentation to find out how to work with the project. 

## Prerequisites

Use the following tools to set up the project:

* Version 1.11 or higher of [Go](https://golang.org/dl/)
* Version 0.5.1 or higher of [Dep](https://github.com/golang/dep)
* Version 2.0.0 of [Kustomize](https://github.com/kubernetes-sigs/kustomize)
* Version 1.0.7 of [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Mockery](https://github.com/vektra/mockery) 

## Available controllers

### Notifier
Notifier is responsible for watching all ProwJobs and send alerts on Slack channel.
For information about the available configuration, see the **Use environment variables** section. 

The ProwJob can be ignored by adding label **prow.k8s.io/slack.skipReport** with value `true` under the ProwJob **metadata** entry. 
```yaml
apiVersion: prow.k8s.io/v1
kind: ProwJob
metadata:
  labels:
    prow.k8s.io/slack.skipReport: true # this job will be ignored by Slack reporter
  name: 47291cd0-2bb4-11e9-9e45-0a580a2c0027
  namespace: default
spec:
  job: post-master-kyma-gke-upgrade
  type: postsubmit
  ...
```

## Usage

### Run a local version

To run the controller outside the cluster, run this command:

```bash
make run
```

### Build a production version

To build the production Docker image, run this command:

```bash
IMG={image_name}:{image_tag} make docker-build
```

The variables are:

* `{image_name}` which is the name of the output image.
* `{image_tag}` which is the tag of the output image.

### Use environment variables
Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|-----|---------|--------|------------|
| **NOTIFIER_GITHUB_ACCESS_TOKEN** | Yes |  | The GitHub token for querying GitHub api. |
| **NOTIFIER_SLACK_TOKEN** | Yes | | The Slack token used or publishing messages to Slack channel. Find more information [here](https://api.slack.com/docs/token-types#bot). |
| **NOTIFIER_SLACK_REPORTER_CHANNEL** | Yes |  | The Slack channel name where notification are posted. |
| **NOTIFIER_SLACK_REPORTER_ACT_ON_PROW_JOB_TYPE** | No | `periodic;postsubmit` | The names of the ProwJob types you want to observe. Multiple type names should be separated by comma or semicolon.
| **NOTIFIER_SLACK_REPORTER_ACT_ON_PROW_JOB_STATE** | No | `failure;error` | The names of the ProwJob states you want to observe. Multiple state names should be separated by comma or semicolon. 

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:
```bash
make resolve
```

### Run tests

To run all unit tests, use the following command:

```bash
make test
```
