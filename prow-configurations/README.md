# Prow Installation

## Overview

This folder contains the installation script and the set of configurations for Prow. 

## Prerequisites

Install the following tools:

- Kubernetes 1.10+ on GKE
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- OpenSSL

### Provision cluster
Follow k8s/prow documentation steps to provision cluster on GKE (https://github.com/kubernetes/test-infra/blob/master/prow/getting_started.md#create-the-cluster).

## Installation
Ensure that kubectl is pointing to correct cluster. For GKE execute following command:
```
gcloud container clusters get-credentials {CLUSTER_NAME} --zone={ZONE_NAME} --project={PROJECT_NAME}
```

Set an OAuth2 token that has read and write access to the bot account. You can set it either as an environment variable named `OAUTH` or interactively during the installation. 
To generate new token, on Github page select given account --> Settings --> Developer Settings --> Personal Access Token --> Generate New Token. 
In the new window select all scopes and click `Generate token`. 
> Note: We recommend to create separate account instead of using your personal account. 

Run the following script to start the installation process: 

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

To check if installation was successful, perform following steps:
- check if all pods are up and running:
```kubeclt get pods```
- check if deck is accessible from outside of the cluster
```kubectl get ing ing```
Copy address and open it in a browser. Dashboard with Prow status should be displayed.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```bash
./remove-prow.sh
```