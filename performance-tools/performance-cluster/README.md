# Performance Cluster

## Overview

Are a set of scripts that deploy on demand a kyma cluster on GCP.

## Commands

- `action`: It is a required command which indicates the action to be executed for the scripts. Possible action values are `create` or `delete`
- `cluster-grade`: It is a required command which indicates the cluster grade of the kyma cluster. Possible action values are `production` or `development`


## Usage

Creates GKE cluster and install a  Kyma:

- cluster grade development

```bash

./cluster.sh --action create --cluster-grade development

```

- cluster grade production

Creates GKE cluster and install a  Kyma cluster grade production:

```bash

./cluster.sh --action create --cluster-grade production

```

Delete Kyma and remove GKE cluster:

- cluster grade development

```bash
./cluster.sh --action delete --cluster-grade development
```

- cluster grade production

```bash
./cluster.sh --action delete --cluster-grade production
```

## Expected environment variables:

- DOCKER_REGESTRY
- DOCKER_PUSH_REPOSITORY - Docker repository hostname. Ex. ""
- DOCKER_PUSH_DIRECTORY - Docker "top-level" directory (with leading "/")
- KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation.
   Ex. "/home/${USER}/go/src/github.com/kyma-project"

- CLOUDSDK_CORE_PROJECT - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
- CLOUDSDK_COMPUTE_REGION - GCP compute region. Ex. "europe-west3"
- CLOUDSDK_COMPUTE_ZONE Ex. "europe-west3-a"
- CLOUDSDK_DNS_ZONE_NAME - GCP zone name (not its DNS name!). Ex. ""build-kyma""
- GOOGLE_APPLICATION_CREDENTIALS - GCP Service Account key file path.
  Ex. "/etc/credentials/sa-gke-kyma-integration/service-account.json"

- MACHINE_TYPE (optional): GKE machine type
- CLUSTER_VERSION (optional): GKE cluster version
- INPUT_CLUSTER_NAME - name for the new cluster

### Permissions: 

In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
- Compute Admin
- Kubernetes Engine Admin
- Kubernetes Engine Cluster Admin
- DNS Administrator
- Service Account User
- Storage Admin
- Compute Network Admin

> **NOTE**: Docker container regestry credentials are needed for executing `docker push`. [Authentication methods](https://cloud.google.com/container-registry/docs/advanced-authentication)

