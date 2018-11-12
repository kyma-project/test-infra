# Production Cluster Configuration

## Overview

This instruction provides the steps required to deploy the production cluster for Prow.

## Prerequisites

Use the following tools and configuration:

- Kubernetes 1.10+ on GKE
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to communicate with Kubernetes.
- [gcloud](https://cloud.google.com/sdk/gcloud/) to communicate with Google Cloud Platform.
- The `kyma-bot` GitHub account
- [Kubernetes cluster](./prow-installation-on-forks.md/#provision-a-cluster)
- Two Secrets in the Kubernetes cluster:
    - `hmac-token` which is a Prow HMAC token used for validating GitHub webhooks.
    - `oauth-token` which is a GitHub token with read and write access to the `kyma-bot` account.
- Google Cloud Platform configuration that includes:
    - A [global static IP address](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address) with the `prow-production` name.
    - A [DNS registry](https://cloud.google.com/dns/docs/quickstart#create_a_managed_public_zone) for the `status.build.kyma-project.io` domain that points to the `prow-production` address.

## Installation

1. Export these environment variables, where:
      - **BUCKET_NAME** is a GCS bucket in the Google Cloud project that is used to store Prow Secrets.
      - **KEYRING_NAME** is the KMS key ring.
      - **ENCRYPTION_KEY_NAME** is the key name in the key ring that is used for data encryption.

      ```
      export BUCKET_NAME=kyma-prow
      export KEYRING_NAME=kyma-prow
      export ENCRYPTION_KEY_NAME=kyma-prow-encryption
      ```

2. Run the following script to start the installation process:

```bash
./install-prow.sh
```

The installation script performs the following steps to install Prow:

- Deploy the NGINX Ingress Controller.
- Create a ClusterRoleBinding.
- Deploy Prow components with the `a202e595a33ac92ab503f913f2d710efabd3de21`revision.
- Deploy Cert Manager.
- Deploy Secure Ingress.
- Removes Insecure Ingress.

3. Verify the installation.

To check if the installation is successful, perform the following steps:

- Check if all Pods are up and running:
   `kubeclt get pods`
- Check if the Deck is accessible from outside of the cluster:
   `kubectl get ingress tls-ing`
   Copy the address of the `tls-ing` Ingress and open it in a browser to display the Prow status on the dashboard.
