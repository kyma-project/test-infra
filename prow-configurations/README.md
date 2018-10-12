# Prow Installation

## Overview

This folder contains the installation script and the set of configurations for Prow. 

## Prerequisites

Install the following tools:

- Kubernetes 1.10+ on GKE
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- OpenSSL

## Installation

Set an OAuth2 token that has read and write access to the bot account. You can set it either as an environment variable named `OAUTH` or interactively during the installation.

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
- Upload the set of configurations for plugins.

### Cleanup

To clean up everything created by the installation script, run the removal script:

```bash
./remove-prow.sh
```