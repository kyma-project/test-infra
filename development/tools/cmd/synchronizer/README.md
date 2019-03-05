# Component synchronizer

## Overview

This command walks recursively through all Kyma repository components and searches makefile command which returns localization of component version. 
Next compares component version with actual component git hash commit. If component is out of date then an appropriate warning will be send to slack channel.

## Usage

The command is run periodically by prow job. To run script localy set below environment variable and run as regular go application
ensuring earlier all vendor dependencies are up to date.
Run from `/development/tools`
```
dep ensure -v -vendor-only
```
then run from current directory
```
go run main.go
```

### Environment variables

| Name                                  | Required  | Description                              |
| :------------------------------------ | :------:  | :--------------------------------------- |
| **KYMA_PROJECT_DIR**                  |    Yes    | Path to the `kyma-project` directory which contains the Kyma source code                  |
| **SLACK_CLIENT_TOKEN**                |    Yes    | Token to slack where message about out of date components will be sent                    |
| **STABILITY_SLACK_CLIENT_CHANNEL_ID** |    Yes    | Channel name where message about out of date components will be sent                      |
| **OUT_OF_DATE_DAYS**                  |    No     | Number of days after which component will be treated as out of date (default value is 3)  |
