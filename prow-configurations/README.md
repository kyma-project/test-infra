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
Follow [this](https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md#create-the-cluster) Prow documentation steps to provision a new cluster on GKE.

## Installation
1. Ensure that kubectl points to the correct cluster. For GKE, execute the following command:
```
gcloud container clusters get-credentials {CLUSTER_NAME} --zone={ZONE_NAME} --project={PROJECT_NAME}
```

2. Set an OAuth2 token that has the read and write access to the bot account. You can set it either as an environment variable named `OAUTH` or interactively during the installation. 
To generate a new token, go to the **Settings** tab of a given GitHub account and click **Developer Settings**. Choose **Personal Access Token** and **Generate New Token**.
In the new window, select all scopes and click **Generate token**. 
>**NOTE:** It is recommended to create a separate account instead of using your personal one. 

3. Run the following script to start the installation process: 

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
<!--- 
TODO not implemented yet:
- Upload the set of configurations for plugins.
-->

To check if the installation is successful, perform the following steps:
1. Check if all Pods are up and running:
```kubeclt get pods```
2. Check if the Deck is accessible from outside of the cluster:
```kubectl get ingress ing```
Copy the address of the ingress `ing` and open it in a browser to display the Prow status on the dashboard.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```bash
./remove-prow.sh
```