
# Prow Secrets

## Overview

This document lists all types of Secrets used in the `kyma-prow` and `workload-kyma-prow` clusters, where all the ProwJobs are executed.
>**NOTE:** All Secrets are stored in the Google Cloud Storage (GCS) bucket.


## kyma-prow cluster

| Prow Secret   | Description | 
| :---------- | :---------------- | 
| **hmac-token**| Used for validating GitHub webhooks. It is manually generated using the `openssl rand -hex 20` command.| 
| **oauth-token**| Personal access token called `prow-production` used by the `kyma-bot` GitHub user. | 
|**sap-slack-bot-token**| Token for publishing messages to the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot). |
|**workload-cluster**| Stores the `workload-kyma-prow` cluster certificate. |
| **slack-token** | OAuth token for the Slack bot user. It is used by Crier. |

## workload-kyma-prow cluster

| Secret   | Description | 
| :---------- | :---------------- | 
| **whitesource** keys | Copied directly from the bucket when executing a job. These Secrets are not stored on the cluster. | 
| **github-integration** | Used to authorize GitHub applications configured in the `kyma-project` organization. See the **[OAuth Apps](https://developer.github.com/apps/building-oauth-apps/)** section in GitHub.|
| **sa-*** | Service Accounts used in pipelines. Find more information [here](/docs/prow/authorization.md).| 
| **kyma-website-bot-zenhub-token** | ZenHub token for the `kyma-website-bot` account.| 
| **kyma-bot-github-token**| Personal access token called `prow-job` used by the `kyma-bot` GitHub user.| 
| **kyma-guard-bot-github-token** | Personal access token for the `kyma-guard-bot` GitHub account.| 
| **kyma-bot@sap.com**| Stores credentials to the `kyma-bot` GitHub account. |
| **kyma-bot-npm-token** | Token for publishing npm packages in the `npmjs.com` registry. The `kyma-bot` user credentials are used to authenticate to the registry. The Secret is used by the `post-master-varkes` ProwJob. |
| **gardener-kyma-prow-kubeconfig** | Kubeconfig file that allows connection to the `kyma-prow` Gardener project.| 
| **slack-nightly-token**| Token that allows the stability checker to push notifications to Slack. | 
| **sap-slack-bot-token** | Token for publishing messages to the SAP CX workspace. Find more information [here](https://api.slack.com/docs/token-types#bot).|
| **kyma-alerts-slack-api-url** | Token for publishing messages to the SAP CX workspace.  It is used by nightly and weekly ProwJobs.|
| **neighbors-alerts-slack-api-url** | Publishes alerts to the private `neighbors` Slack channel.|
| **kyma-azure-credential-*** | Azure subscription and service principal credentials. |
| **kyma-snyk-token** | Token that allows authentication to Snyk CLI. It is used by the `vulnerability-scanner` ProwJob. |
| **kyma-website-bot-*** | Personal access token of the `kyma-website-bot` GitHub account. It is responsible for publishing the `kyma-project.io` website. |
