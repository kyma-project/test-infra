# Overview

The `issueLabeled` Function is executed by Kyma [Serverless](https://kyma-project.io/docs/components/serverless/) when the registered **issuesevent.labeled** event occurs. All **issuesevent.labeled** events for the `internal-incident` or `customer-incident` labels will be processed by this Function.

## Prerequisites

You need Kyma Runtime to run this Function.

## Installation

Push the [handler.py](handler.py) and [requirements.txt](requirements.txt) files to your GitHub repository and [point Serverless](https://kyma-project.io/docs/components/serverless/#tutorials-create-a-function-from-git-repository-sources) to download its code for building a Function image.

## Configuration

Function is using Service Binding Usage to get Slack API URL. To point function to correct environment variable, set its name in function code.

```
client = WebClient(base_url="{}/".format(os.environ['KYMA_SLACK_SLACK_CONNECTOR_85DED56E_303B_43B3_A950_8B1C3D519561_GATEWAY_URL']))
```
