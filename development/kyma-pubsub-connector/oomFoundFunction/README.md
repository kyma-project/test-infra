# Overview

The oomFound Function is executed by Kyma [Serverless](https://kyma-project.io/docs/components/serverless/) when a registered **oomevent.found** event occurs. The Function is written in Python.

## Prerequisites

You need the Kyma Runtime to run this Function.

## Installation

1. Push the [handler.py](handler.py) and [requirements.txt](requirements.txt) files to your GitHub repository 
2. Point Serverless to [download the repository's code](https://kyma-project.io/docs/components/serverless/#tutorials-create-a-function-from-git-repository-sources) which is needed for building a Function image.

## Configuration

| Environment variable name | Required | Description |
|----------------|----------|-------------|
| **SLACK_API_ID** | Yes | Slack application API ID. |
| **SLACK_BOT_TOKEN** | Yes | Slack application bot token. |
| **NOTIFICATION_SLACK_CHANNEL** | Yes | Target Slack channel name. |

The oomFound Function uses [ServiceBindingUsage](https://kyma-project.io/docs/components/serverless/#tutorials-bind-a-service-instance-to-a-function) to get the Slack API URL. To point the Function to the environment variable provided by the ServiceBindingUsage, set the **SLACK_API_ID** environment variable by providing its value in the *slackConnector.apiId* key in [values.yaml](../pubsubConnector/values.yaml). The Slack API ID will be used to build the target [variable name](https://github.com/kyma-project/test-infra/blob/5f6e98dd0cf692156a0d1caf9b3df5c27d39368d/development/pubsub-connector/oomFoundFunction/handler.py#L15).

The Function uses the **SLACK_BOT_TOKEN** environment variable to authorise sending messages to Slack channels. To use a valid token, set it on chart installation with helm command argument `--set slackConnector.botToken="xoxb-xxx-xxxx-xxxx-xxxx-xxx"` or set it in [values.yaml](../pubsubConnector/values.yaml).

To send your messages to chosen channel, provide its name in the **NOTIFICATION_SLACK_CHANNEL** environment variable. This variable is set by the *function.oomevent.found.notificationSlackChannel* key in [values.yaml](../pubsubConnector/values.yaml).

## Development

This is a sample data payload delivered to the Function:

```
{
    'ID': '2401353396977684',
    'Data': 'eyJwcm9qZWN0Ijoic2FwLWt5bWEtcHJvdyIsInRvcGljIjoicHJvd2pvYnMiLCJydW5pZCI6Imt5bWEtY2xpLWFscGhhLXVuaW5zdGFsbC1na2UiLCJzdGF0dXMiOiJmYWlsdXJlIiwidXJsIjoiaHR0cHM6Ly9zdGF0dXMuYnVpbGQua3ltYS1wcm9qZWN0LmlvL3ZpZXcvZ3Mva3ltYS1wcm93LWxvZ3MvbG9ncy9reW1hLWNsaS1hbHBoYS11bmluc3RhbGwtZ2tlLzEzOTQzNjcyNjA2OTY1MTQ1NjAiLCJnY3NfcGF0aCI6ImdzOi8va3ltYS1wcm93LWxvZ3MvbG9ncy9reW1hLWNsaS1hbHBoYS11bmluc3RhbGwtZ2tlLzEzOTQzNjcyNjA2OTY1MTQ1NjAiLCJyZWZzIjpbeyJvcmciOiJreW1hLXByb2plY3QiLCJyZXBvIjoiY2xpIiwiYmFzZV9yZWYiOiJtYWluIiwicGF0aF9hbGlhcyI6ImdpdGh1Yi5jb20va3ltYS1wcm9qZWN0L2NsaSJ9LHsib3JnIjoia3ltYS1wcm9qZWN0IiwicmVwbyI6Imt5bWEiLCJiYXNlX3JlZiI6Im1haW4iLCJwYXRoX2FsaWFzIjoiZ2l0aHViLmNvbS9reW1hLXByb2plY3Qva3ltYSJ9LHsib3JnIjoia3ltYS1wcm9qZWN0IiwicmVwbyI6InRlc3QtaW5mcmEiLCJiYXNlX3JlZiI6Im1haW4iLCJwYXRoX2FsaWFzIjoiZ2l0aHViLmNvbS9reW1hLXByb2plY3QvdGVzdC1pbmZyYSJ9XSwiam9iX3R5cGUiOiJwZXJpb2RpYyIsImpvYl9uYW1lIjoia3ltYS1jbGktYWxwaGEtdW5pbnN0YWxsLWdrZSJ9',
    'Attributes': None,
    'PublishTime': '2021-05-17T19:22:27.185Z',
    'DeliveryAttempt': None,
    'OrderingKey': ''
}
```

This is an example of a base64 decoded Data property:
```
{
    'project': 'sap-kyma-prow',
    'topic': 'prowjobs',
    'runid': 'kyma-cli-alpha-uninstall-gke',
    'status': 'failure',
    'url': 'https://status.build.kyma-project.io/view/gs/kyma-prow-logs/logs/kyma-cli-alpha-uninstall-gke/1394367260696514560',
    'gcs_path': 'gs://kyma-prow-logs/logs/kyma-cli-alpha-uninstall-gke/1394367260696514560',
    'refs': [
        {
            'org': 'kyma-project',
            'repo': 'cli',
            'base_ref': 'main',
            'path_alias': 'github.com/kyma-project/cli'
        },
        {
            'org': 'kyma-project',
            'repo': 'kyma',
            'base_ref': 'main',
            'path_alias': 'github.com/kyma-project/kyma'
        },
        {
            'org': 'kyma-project',
            'repo': 'test-infra',
            'base_ref': 'main',
            'path_alias': 'github.com/kyma-project/test-infra'
        }
    ],
    'job_type': 'periodic',
    'job_name': 'kyma-cli-alpha-uninstall-gke'
}
```
