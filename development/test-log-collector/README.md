# Test Log Collector

## Overview

The purpose of the Test Log Collector is to gather logs from testing pods and to send them to appropriate slack channels.
It is intended to be run after Octopus ClusterTestSuite, as it uses pods labels created by Octopus.

## Usage

To use the test-log-collector, navigate to its chart directory and run it with appropriate parameters. See the example

```bash
helm install test-log-collector \
--namespace kyma-system ./chart/test-log-collector \
--set slackToken=${slack_token} \
--wait \
--timeout 600s
```

## Configuration

Test Log Collector dispatches logs to particular Slack channels based on configuration file, located [here](./chart/test-log-collector/files/config.yaml). It has a form of a list, where each element has to have such fields:

- testCases
- channelName
- channelID
- onlyReportFailure

| Parameter             | Description                                                                                                                                                                            |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **testCases**         | List of test names, from which logs should be sent to particular channel. Specifying `default` item in that list creates a sink for test cases, which have unspecified target channel. |
| **channelName**       | Name of the channel, to which logs are sent. **Must** start with "#"                                                                                                                   |
| **channelID**         | ID of the channel, to which logs are sent                                                                                                                                              |
| **onlyReportFailure** | Indicates whether only logs from failed tests should be sent to Slack channel                                                                                                          |

You can get available test cases by running `kyma test definitions`.

`channelID` can be obtained by right clicking channel in Slack and choosing `Copy link` option. Channel ID is the last part of that link. If the channel link is https://example.slack.com/archives/CPBNQ4KNG, then the ID equals CPBNQ4KNG.

See example configuration:

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

Based on that configuration Test Log Collector will:

- send logs from failed `serverless` and `rafter` test cases to `#serverless-core-channel`
- send logs from remaining failed **and** successful test cases to `#default-msg-channel`
