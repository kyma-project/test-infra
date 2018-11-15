# Production Cluster Configuration

## Overview

This instruction provides the steps required to deploy a production cluster for Prow.

## Prerequisites

Use the following tools and configuration:

- Kubernetes 1.10+ on Google Kubernetes Engine (GKE)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to communicate with Kubernetes
- [gcloud](https://cloud.google.com/sdk/gcloud/) to communicate with Google Cloud Platform (GCP)
- The `kyma-bot` GitHub account
- [Kubernetes cluster](./prow-installation-on-forks.md/#provision-a-cluster)
- Two Secrets in the Kubernetes cluster:
    - `hmac-token` which is a Prow HMAC token used to validate GitHub webhooks
    - `oauth-token` which is a GitHub token with read and write access to the `kyma-bot` account
- Two buckets on Google Cloud Storage (GCS), one for storing Secrets and the second for storing logs
- GCP configuration that includes:
    - A [global static IP address](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address) with the `prow-production` name
    - A [DNS registry](https://cloud.google.com/dns/docs/quickstart#create_a_managed_public_zone) for the `status.build.kyma-project.io` domain that points to the `prow-production` address

## Installation

1. When you communicate for the first time with Google Cloud, set the context to your Google Cloud project.

Export the **PROJECT** variable and run this command:

```
gcloud config set project $PROJECT
```

2. Make sure that kubectl points to the correct cluster.

Export these variables:  

```
export CLUSTER_NAME=prow-production
export ZONE=europe-west3-b
export PROJECT=kyma-project
```

For GKE, run the following command:

```
gcloud container clusters get-credentials $CLUSTER_NAME --zone=$ZONE --project=$PROJECT
```
3. Export these environment variables, where:
  - **BUCKET_NAME** is a GCS bucket in the Google Cloud project that stores Prow Secrets.
  - **KEYRING_NAME** is the KMS key ring.
  - **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.

```
export BUCKET_NAME=kyma-prow
export KEYRING_NAME=kyma-prow
export ENCRYPTION_KEY_NAME=kyma-prow-encryption
```

4. Run the following script to start the installation process:

```bash
./install-prow.sh
```

The installation script performs the following steps to install Prow:

- Deploy the NGINX Ingress Controller.
- Create a ClusterRoleBinding.
- Deploy Prow components with the `a202e595a33ac92ab503f913f2d710efabd3de21`revision.
- Deploy the Cert Manager.
- Deploy secure Ingress.
- Remove insecure Ingress.

5. Verify the installation.

To check if the installation is successful, perform the following steps:

- Check if all Pods are up and running:
   `kubeclt get pods`
- Check if the Deck is accessible from outside of the cluster:
   `kubectl get ingress tls-ing`
- Copy the address of the `tls-ing` Ingress and open it in a browser to display the Prow status on the dashboard.
