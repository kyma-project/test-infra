# Prow Installation

## Overview

This folder contains the installation script and the set of configurations for Prow.

>**NOTE:** The following instructions assume that you are signed in to the Google Cloud project with administrative rights.

## Prerequisites

Install the following tools:

- Kubernetes 1.10+ on GKE
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- OpenSSL

### Provision a cluster
Use the `provision-cluster.sh` script or follow [these](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md#create-the-cluster) instructions to provision a new cluster on GKE. Ensure that kubectl points to the correct cluster. For GKE, execute the following command:

```
gcloud container clusters get-credentials {CLUSTER_NAME} --zone={ZONE_NAME} --project={PROJECT_NAME}
```

## Installation

1. Set an OAuth2 token that has the read and write access to the bot account. You can set it either as an environment variable named `OAUTH` or interactively during the installation.
To generate a new token, go to the **Settings** tab of a given GitHub account and click **Developer Settings**. Choose **Personal Access Token** and **Generate New Token**.
In the new window, select all scopes and click **Generate token**.
>**NOTE:** Create a separate bot account instead of using your personal one. If the Prow bot account is the same as account that creates a job-triggering comment, the job is not triggered!

2. Run the following script to start the installation process:

```bash
./install-prow.sh
```

The installation script performs the following steps to install Prow:

- Deploy the NGINX Ingress Controller.
- Create a ClusterRoleBinding.
- Create a HMAC token to be used for GitHub Webhooks.
- Create secrets for HMAC and OAuth2 to be used by Prow.
- Deploy Prow components using the `starter.yaml` file from the `prow/cluster` directory.
- Add annotations for the Prow Ingress to make it work with the NGINX Ingress Controller.

To check if the installation is successful, perform the following steps:
1. Check if all Pods are up and running:
`kubeclt get pods`
2. Check if the Deck is accessible from outside of the cluster:
`kubectl get ingress ing`
Copy the address of the ingress `ing` and open it in a browser to display the Prow status on the dashboard.

## Configuration
To allow sending events from Github repository to Prow, configure Webhook as described [here](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md#add-the-webhook-to-github).
You can configure Prow by specifying the `plugins.yaml` and `config.yaml` files. To generate them, use `./development/generate.sh`. The `generate.sh` script combines the `plugins.yaml.tpl` and `config.yaml.tpl` template files with actual values provided as a JSON file and generates output to the `plugins.yaml` and `config.yaml` files. The following snippet is an exmample of the JSON file:

```
{
  "OrganizationOrUser":"yourgithubuser"
}
```

Store the updated `plugins.yaml` and `config.yaml` files, which are available for a cluster, in the `prow` directory.

>**NOTE:** You can provide a path to the JSON file from the console input or by specifying the `INPUT_JSON` environment variable.

To check if the `plugins.yaml` and `config.yaml` configuration files are correct, run the `development/check.sh` script.
In case of changes in the plugins configuration, use the `update-plugins.sh` to apply changes on a cluster.
In case of changes in the jobs configuration, use the `update-config.sh` to apply changes on a cluster.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```bash
./remove-prow.sh
```
