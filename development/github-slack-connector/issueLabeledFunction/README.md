# Overview

Function issueLabeled is executed by Kyma [Serverless](https://kyma-project.io/docs/components/serverless/) when registered issuesevent.labeled event occur.

## Prerequisites

1. You need Kyma Runtime to run this function.

## Installation

Push [handler.py](handler.py) and [requirements.txt](requirements.txt) files to Github repository and [point serverless](https://kyma-project.io/docs/components/serverless/#tutorials-create-a-function-from-git-repository-sources) to download its code for building function image.

## Configuration

Function is using [ServiceBindingUsage](https://kyma-project.io/docs/components/serverless/#tutorials-bind-a-service-instance-to-a-function) to get Slack API URL. To point function to correct environment variable set its name in function code.

```
client = WebClient(base_url="{}/".format(os.environ['KYMA_SLACK_SLACK_CONNECTOR_85DED56E_303B_43B3_A950_8B1C3D519561_GATEWAY_URL']))
```
