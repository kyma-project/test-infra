# Cluster Integration Job

## Overview

The folder contains the source code for the integration job that installs and tests Kyma on a temporary cluster provisioned on Google Kubernetes Engine (GKE).
This job runs as a pre-submit job for pull requests.

### Pipeline logic

The integration job is a pipeline that consists of multiple steps:
- Build a Kyma-Installer image.
- Provision a GKE cluster.
- Reserve an IP address.
- Create DNS Entry for reserved IP Address
- Generate TLS Certificate
- Install Kyma on the GKE cluster
- Test Kyma installation
- Ensure to clean up all provisioned resources, also in case of an error.

### Project structure

The main entry point of the entire pipeline is `pr-gke-integration.sh` script that invokes other helper scripts / CLI tools.
The Pipeline uses a toolset from `Bootstrap` image as defined in this repository.

### Required environment variables

This script takes its input configuration from environment variables.
The following environment variables are required:

- `REPO_OWNER` - Set up by prow, repository owner/organization
- `REPO_NAME` - Set up by prow, repository name
- `PULL_NUMBER` - Set up by prow, Pull request number
- `DOCKER_PUSH_REPOSITORY` - Docker repository hostname
- `DOCKER_PUSH_DIRECTORY` - Docker "top-level" directory (with leading "/")
- `KYMA_PROJECT_DIR` - directory path with Kyma sources to use for installation
- `CLOUDSDK_CORE_PROJECT` - GCP project for all GCP resources used during execution (Service Account, IP Address, DNS Zone, image registry etc.)
- **CLOUDSDK_COMPUTE_REGION** is a GCP compute region.
- **CLOUDSDK_DNS_ZONE_NAME** is a GCP zone name which is different from the DNS name.
- **GOOGLE_APPLICATION_CREDENTIALS** is the path to the GCP service account key file.

### Required permissions

The pipeline accesses GCP using a service account configured with the **GOOGLE_APPLICATION_CREDENTIALS** environment variable.
This service account must have GCP permissions equivalent to the following GCP roles:

- Compute Network Admin (`roles/compute.networkAdmin`)
- Kubernetes Engine Admin (`roles/container.admin`)
- Kubernetes Engine Cluster Admin (`roles/container.clusterAdmin`)
- DNS Administrator (`roles/dns.admin`)
- Service Account User (`roles/iam.serviceAccountUser`)
- Storage Admin (`roles/storage.admin`)
