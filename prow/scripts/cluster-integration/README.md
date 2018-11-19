# Cluster Integration Job

## Overview

The folder contains the source code for the integration job that installs and tests Kyma on a temporary cluster provisioned on Google Kubernetes Engine (GKE).
For now, it's only executed for Pull Requests as a pre-submit job.

### Pipeline logic

The integration job is a pipeline that consists of multiple steps. Their order is not strict and some can run in parallel:
- Build Kyma-Installer Image
- Provision GKE cluster
- Reserve IP Address
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
- `CLOUDSDK_COMPUTE_REGION` - GCP compute region
- `CLOUDSDK_DNS_ZONE_NAME` -GCP zone name (not its DNS name!)
- `GOOGLE_APPLICATION_CREDENTIALS` - GCP Service Account key file path

### Required permissions

The pipeline authorizes to GCP as a service account configured with `GOOGLE_APPLICATION_CREDENTIALS` environment variable.
This service account must have GCP permissions equivalent to the following GCP roles:

- `Compute Network Admin`
- `Kubernetes Engine Admin`
- `Kubernetes Engine Cluster Admin`
- `DNS Administrator`
- `Service Account User`
- `Storage Admin`

