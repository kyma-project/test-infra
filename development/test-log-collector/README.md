# Test Log Collector

## Overview

The purpose of the Test Log Collector is to gather logs from the testing Pods and to send them to the appropriate Slack channels.
It is intended to be run after the Octopus ClusterTestSuite, as it uses Pods labels from Octopus. By design, the Test Log Collector sends logs only from the newest, finalized ClusterTestSuite.

## Prerequisites

To send the message to any Slack channel, you need to add the [Slack app](https://api.slack.com/start) to that channel and have its token. Slack app tokens typically have the `xoxb-` prefix. The Slack app must have the following bot token scopes:

- `channels:history`
- `chat:write`
- `files:write`

## Usage

To use the Test Log Collector, navigate to its chart directory, and run it with appropriate parameters. See the example:

```bash
helm install test-log-collector \
--namespace kyma-system ./chart/test-log-collector \
--set slackToken=${slack_token} \
--wait \
--timeout 600s
```

### Configuration

The Test Log Collector dispatches logs to particular Slack channels based on the [configuration file](./chart/test-log-collector/files/config.yaml). This config file is a list which uses the following fields to configure the application:

| Parameter             | Description                                                                                                                                                                            |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **testCases**         | List of test names from which logs should be sent to a particular channel. Specifying the `default` item in that list creates a sink for test cases which have an unspecified target channel. |
| **channelName**       | Name of the channel to which logs are sent. It must start with "#".                                                                                                                   |
| **channelID**         | ID of the channel to which logs are sent                                                                                                                                              |
| **onlyReportFailure** | Parameter that indicates whether only logs from failed tests should be sent to a Slack channel                                                                                                         |

You can get available test cases by running `kyma test definitions`.

You can obtain `channelID` by right-clicking the channel on Slack and choosing the **Copy link** option. `channelID` is the last part of that link. If the channel link is `https://example.slack.com/archives/CPBNQ4KNG`, the ID equals `CPBNQ4KNG`.

See the example configuration file:

```yaml
- channelName: "#default-msg-channel"
  channelID: "some-channel-id"
  onlyReportFailure: false
  testCases:
    - default
- channelName: "#serverless-core-channel"
  channelID: "some-other-channel-id2"
  onlyReportFailure: true
  testCases:
    - serverless
    - rafter
```

Based on that configuration, the Test Log Collector:

- Sends logs from failed `serverless` and `rafter` test cases to `#serverless-core-channel`.
- Sends logs from the remaining failed and successful test cases to `#default-msg-channel`.
