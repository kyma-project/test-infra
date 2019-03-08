# Component synchronizer

## Overview

This command walks recursively through all Kyma repository components and searches makefile command which returns localization of the charts where the given component is used. 
Next compares component version with actual component git hash commit. If component is out of date then an appropriate warning will be displayed in logs.
It is also possible to send alert message to other applications (currently available application is slack). 

## Usage

To run script localy set below environment variables and run as regular go application
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
| **KYMA_PROJECT_DIR**                  |    Yes    | Path to the `kyma-project/kyma` directory which contains the Kyma source code. |
| **SLACK_CLIENT_TOKEN**                |    No     | Token to slack where message about out of date components will be sent. If not set message will be not set, all information will be available in logs. |
| **STABILITY_SLACK_CLIENT_CHANNEL_ID** |    No     | Channel name where message about out of date components will be sent. If not set message will be not set, all information will be available in logs. |
| **OUT_OF_DATE_DAYS**                  |    No     | Number of days after which component will be treated as out of date (default value is 3). |
