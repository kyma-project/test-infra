# Prow Production Configuration

## Prerequisites

### Tools

Install the following tools:

- Kubernetes 1.10+ on GKE
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [gcloud](https://cloud.google.com/sdk/gcloud/)

### Google Cloud Platform configuration

- A global static IP address with the `prow-production` name.
- A DNS registry for the `status.build.kyma-project.io` domain that points to the `prow-production` address.

### Secrets:

- hmac-token - Prow HMAC token used for GitHub Webhooks
- oauth-token - GitHub token with read/write access to `kyma-bot` account
- compute-service-account - Google Cloud Service Account with roles:
  - Service Account User
  - Compute Admin
  - Compute OS Admin Login

## Installation

1. Ensure that `kubectl` points to the correct cluster. For GKE, execute the following command:

```
gcloud container clusters get-credentials {CLUSTER_NAME} --zone={ZONE_NAME} --project={PROJECT_NAME}
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

To check if the installation is successful, perform the following steps:

1. Check if all Pods are up and running:
   `kubeclt get pods`
2. Check if the Deck is accessible from outside of the cluster:
   `kubectl get ingress tls-ing`
   Copy the address of the ingress `tls-ing` and open it in a browser to display the Prow status on the dashboard.
