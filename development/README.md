# Prow Installation

## Overview

This folder contains the installation script and the set of configurations for Prow.

> **NOTE:** The following instructions assume that you are signed in to the Google Cloud project with administrative rights.

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

   > **NOTE:** Create a separate bot account instead of using your personal one. If the Prow bot account is the same as the account that creates a job-triggering comment, the job is not triggered.

2. Export the environment variables:

- **BUCKET_NAME** is a GCS bucket in the Google Cloud project that is used to store Prow Secrets.
- **KEYRING_NAME** is the KMS key ring.
- **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.

> **NOTE:** For more details on setting up Google Cloud Project refer to [Prow Secrets management](https://github.com/kyma-project/test-infra/blob/master/docs/prow-secrets-management.md)

The Prow installation script assumes that the following service accounts exist:

- **sa-gke-kyma-integration** with a role that allows the account to create Kubernetes clusters.
- **sa-vm-kyma-integration** with roles that allow the account to provision virtual machines.
- **sa-gcs-plank** with roles that allow the account to store objects in a Bucket.
- **sa-gcr-push** with roles that allow the account to push images to Google Container Repository.

The account files are encrypted with the **ENCRYPTION_KEY_NAME** key from **KEYRING_NAME** and are stored in **BUCKET_NAME**.

3. Run the following script to start the installation process:

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

## Development

[Configure Webhook](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md#add-the-webhook-to-github) to enable sending Events from a GitHub repository to Prow.
When you use our script `install-prow.sh` to install Prow on your cluster the list of plugins and configuration is empty. You can configure Prow by specifying the `plugins.yaml` and `config.yaml` files.

To check if the `plugins.yaml` and `config.yaml` configuration files are correct, run the `check.sh` script.
In case of changes in the plugins configuration, use the `update-plugins.sh` to apply changes on a cluster.
In case of changes in the jobs configuration, use the `update-config.sh` to apply changes on a cluster.

### Strategy for organising jobs

The `test-infra` repository has defined configurations for the Prow cluster in the `prow` subdirectory. This directory has the following structure:

- `cluster` directory, which contains all `yaml` files for Prow cluster provisioning
- `jobs/{repository_name}` directory, which contains all files with jobs definitions, each file must have unique name.
- `config.yaml` file, which contains Prow configuration without job definitions
- `plugins.yaml` file, which contains Prow plugins configuration

`jobs/{repository_name}` directories have subdirectories which represent each component and contain job definitions. Job definitions not connected to a particular component, like integration jobs, are defined directly under the `jobs/{repository_name}` directory.

For example:

```
...
prow
|- cluster
| |- starter.yaml
|- jobs
| |- kyma
| | |- components
| | | |- environments
| | | | |- environments.jobs.yaml
| | |- kyma.integration.yaml
|- config.yaml
|- plugins.yaml
...
```

### Convention for naming jobs

When you define jobs for Prow, both **name** and **context** of the job must follow one of these patterns:

- `prow/{repository_name}/{component_name}/{job_name}` for components
- `prow/{repository_name}/{job_name}` for jobs not connected to a particular component

In both cases, `{job_name}` must reflect the job's responsibility.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```bash
./remove-prow.sh
```
