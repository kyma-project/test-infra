# Component synchronizer

## Overview

This script recursively checks all components in the `kyma` repository and searches for Makefiles. In Makefiles, there are targets which return localizations of the charts where the components are used. 
Then,  the script compares components version with the actual components git hash commit. If any component version is out of date, the warning appears in logs.
You can also configure the script to send alert messages to Slack.

## Usage

To run the script locally, set the [environment variables](#environment-variables), navigate to `/development/tools`, and ensure that all vendor dependencies are up to date. Run:
```
dep ensure -v -vendor-only
```
Then, run this command:
```
go run main.go
```

### Environment variables

| Name                                  | Required  | Description                              |
| :------------------------------------ | :------:  | :--------------------------------------- |
| **KYMA_PROJECT_DIR**                  |    YES    | A path to the `kyma-project/kyma` directory which contains the Kyma source code. |
| **SLACK_CLIENT_TOKEN**                |    NO     | A token to the Slack channel where messages about outdated components are sent. If not set, the messages are not sent and the information about the outdated components is available in logs. |
| **STABILITY_SLACK_CLIENT_CHANNEL_ID** |    NO     | An ID of the Slack channel where messages about outdated components are sent. If not set, the messages are not sent and the information about the outdated components is available in logs. |
| **OUT_OF_DATE_DAYS**                  |    NO     | A number of days after which a component is treated as outdated. The default value is `3`. |
