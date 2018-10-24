# Prow Installation

## Overview

This folder contains the installation script and the set of configurations for Prow. 

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

## Secrets management

Some jobs require using sensitive data. You need to encrypt data using KMS and store them in GCP. 

>**NOTE:** Check wether below setup was already done for the {PROJECT_NAME}.

1. Create service account for provisioning GKE cluster for kyma integration tests:
```
gcloud iam service-accounts create {SA_NAME} --display-name {SA_DISPLAY_NAME}
```

2. Create a private key for the {SA_NAME} service account:
```
gcloud iam service-accounts keys create sa-gke-kyma-integration.json --iam-account={SA_NAME}
```

3. Add policy binding for the {SA_NAME} service account:
```
gcloud iam service-accounts add-iam-policy-binding {SA_NAME} --member=serviceAccount:{SA_NAME}@{PROJECT_NAME}.iam.gserviceaccount.com --role={ROLE}
```

4. Create a bucket for storing sensitive data:
```
gsutil mb -p {PROJECT_NAME} gs://{BUCKET_NAME}/
```

5. Create a key ring:
```
gcloud kms keyrings create prow-keyring --location global
```

6. Create a key:
```
gcloud kms keys create prow-encryption-key --location global --keyring prow-keyring --purpose encryption
```

7. Encrypt the secret:
```
gcloud kms encrypt --location global --keyring prow-keyring --key prow-encryption-key --plaintext-file sa-gke-kyma-integration.json --ciphertext-file sa-gke-kyma-integration.json.encrypted
```

8. Upload encrypted secret to GCP:
```
gsutil cp sa-gke-kyma-integration.json.encrypted gs://mst-prow-bucket/
```

9. Delete the `sa-gke-kyma-integration.json` file.

## Installation

1. Set an OAuth2 token that has the read and write access to the bot account. You can set it either as an environment variable named `OAUTH` or interactively during the installation. 
To generate a new token, go to the **Settings** tab of a given GitHub account and click **Developer Settings**. Choose **Personal Access Token** and **Generate New Token**.
In the new window, select all scopes and click **Generate token**. 
>**NOTE:** It is recommended to create a separate account instead of using your personal one. 

2. Run the following script to start the installation process: 

```bash
./install-prow.sh
```

The installation script performs the following steps to install Prow:

- Deploy the NGINX Ingress Controller.
- Create a ClusterRoleBinding.
- Create a HMAC token to be used for GitHub Webhooks.
- Create secrets for HMAC and OAuth2 to be used by Prow.
- Deploy Prow components with the `a202e595a33ac92ab503f913f2d710efabd3de21`revision.
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

>**NOTE:** You can provide a path to the JSON file from the console input or by specifying the `INPUT_JSON` environment variable.

To check if the `plugins.yaml` and `config.yaml` configuration files are correct, run the `development/check.sh` script.
In case of changes in the plugins configuration, use the `update-plugins.sh` to apply changes on a cluster. 
In case of changes in the jobs configuration, use the `update-config.sh` to apply changes on a cluster.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```bash
./remove-prow.sh
```