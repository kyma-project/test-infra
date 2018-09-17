# Prow Installation

## Overview

This folder contains the installation script and the set of configurations for Prow. 

## Prerequisites

- Kubernetes 1.10+ on GKE
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- git
- openssl

## Installation

First, you need to set OAuth2 token that has read and write access to the bot account. You can set it as an environment variable named `OAUTH` or set it interactively during installation.

Run the following script to start the installation process: 

```bash
./install-prow.sh
```

Installation script will accomplish the following steps to install Prow:

- Clone kubernetes/test-infra repo and checkout `a202e595a33ac92ab503f913f2d710efabd3de21` revision.
- Deploy NGINX Ingress Controller.
- Create a ClusterRoleBinding.
- Create a HMAC token to be used for GitHub Webhooks.
- Create secrets for HMAC and OAuth2 to be used by Prow.
- Deploy Prow components.
- Add annotations for Prow Ingress to make it work with NGINX Ingress Controller.
- Change the type of Deck Service to LoadBalancer to access Prow UI (Deck).
- Upload the set of configurations for plugins.
- Remove the test-infra folder.