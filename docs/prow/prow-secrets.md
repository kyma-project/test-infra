
# Prow Secrets

## Overview

This document lists all types of secrets used in kyma-prow cluster as well as workload-kyma-prow cluster, where all the ProwJobs are executed.
>**NOTE:** All secrets are stored in the GCS bucket.


## kyma-prow cluster

| Prow Secret   | Description | 
| :---------- | :---------------- | 
| **hmac-token**| It is used for validating Github webhooks. It is manually generated using `openssl rand -hex 20` command.| 
| **oauth-token**| It stores personal access token (called `Prow - Production`) used by the kyma-bot Github user. | 
|**sap-slack-bot-token**| It is a token for publishing messages to the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot) |
|**workload-cluster**| It stores `workload-kyma-prow` cluster certificate. |
| **slack-token** | It stores oauth token for Slack bot user, it is used by Crier. |

## workload-kyma-prow cluster

| Secret   | Description | 
| :---------- | :---------------- | 
| **whitesource** keys | Are directly copied from the bucket when executing job. Secrets are not stored on the cluster. | 
| **github-integration** | Are used to authorize Github applications configured in kyma-project organization (see `OAuth Apps` section in Github).|
| **sa-*** | Service Accounts used in pipelines. You can find more information [here](/docs/prow/authorization.md).| 
| **kyma-website-bot-zenhub-token** | Stores the ZenHub token for the `kyma-website-bot` account.| 
| **kyma-bot-github-token**| Stores personal access token (called `Prow - Job`) used by the kyma-bot github user.| 
| **kyma-guard-bot-github-token** | Stores personal access token of the kyma-guard-bot Github account.| 
| **kyma-bot@sap.com**| Stores credentials to the kyma-bot Github account. |
| **kyma-bot-npm-token** | It is a token for publishing npm packages in npmjs.com registry. Kyma-bot user credentials are used to authenticate to the registry. The secret is used by `post-master-varkes` ProwJob. |
| **gardener-kyma-prow-kubeconfig** | It is a kubeconfig file, that allows connection to the Gardener `kyma-prow` project.| 
| **slack-nightly-token**| It is a slack token that allows stability checker to push notifications to the Slack. | 
| **sap-slack-bot-token** | It is a token for publishing messages to the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot).|
| **kyma-alerts-slack-api-url** | It is a token for publishing messages to the SAP CX workspace.  It is used by nightly and weekly ProwJobs.|
| **neighbors-alerts-slack-api-url** | Publishes alerts to the private neighbors Slack channel.|
| **kyma-azure-credential-*** | This set of secrets stores Azure subscription and service principal credentials. |
| **kyma-snyk-token** | It is token that allows authentication to Snyk CLI. It is used by `vulnerability-scanner` ProwJob. |
| **kyma-website-bot-*** | Stores a personal access token of the kyma-website-bot Github account. It is responsible for publishing kyma-project.io website. |
