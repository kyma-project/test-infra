# Cluster Integration Job

## Overview

The folder contains the source code for the integration job that installs and tests Kyma on a temporary cluster provisioned on Google Kubernetes Engine (GKE).
This job runs as a pre-submit job for pull requests.

### Pipeline logic

The integration job is a pipeline that consists of multiple steps:
- Build a Kyma-Installer image.
- Provision a GKE cluster.
- Reserve an IP address.
- Create a DNS entry for the reserved IP address.
- Generate a TLS certificate.
- Install Kyma on the GKE cluster.
- Test the Kyma installation.
- Clean up all provisioned resources, also if you get an error.

### Project structure

The structure of the folder looks as follows:

``` 
├── helpers # This directory contains helpers scripts used by pipeline jobs.
│   ├── create-dns-record.sh
│   ├── create-image.sh
│   ├── delete-disks.sh
│   ├── delete-dns-record.sh
│   ├── delete-image.sh
│   ├── deprovision-gke-cluster.sh
│   ├── generate-self-signed-cert.sh
│   ├── provision-gke-cluster.sh
│   ├── release-ip-address.sh
│   └── reserve-ip-address.sh
├── kyma-gke-integration.sh # The purpose of this script is to install and test Kyma on real GKE cluster.
├── kyma-gke-nightly.sh # The purpose of this script is to create from master branch a long-lived GKE cluster. This cluster should be recreated once per day. 
└── kyma-gke-upgradeability.sh # The purpose of this script is to install last Kyma release on GKE cluster, upgrade it with current changes from master/PR/release branch and trigger Kyma testing script.
```

The main entry point for the entire pipeline is the `kyma-gke-integration.sh` script that invokes other helper scripts and CLI tools.

The pipeline uses a toolset from the `Bootstrap` image defined in this repository.

### Required environment variables

This script takes its input configuration from environment variables.
The following environment variables are required:

- **REPO_OWNER** is the repository owner or organization. This variable is set up by Prow.
- **REPO_NAME** is the repository name. This variable is set up by Prow.
- **BUILD_TYPE** is one of pr/master/release. This variable is created by using the `preset-build` label in ProwJob definition
- **DOCKER_PUSH_REPOSITORY** is the Docker repository hostname.
- **DOCKER_PUSH_DIRECTORY** - the Docker top-level directory, preceded by a slash (/).
- **KYMA_PROJECT_DIR** is a directory path with Kyma sources to use for the installation.
- **CLOUDSDK_CORE_PROJECT** is a Google Cloud Platform (GCP) project for all GCP resources used in the script. For example, the resources include service account, an IP address, a DNS Zone, and an image registry.
- **CLOUDSDK_COMPUTE_REGION** is a GCP compute region.
- **CLOUDSDK_DNS_ZONE_NAME** is a GCP zone name which is different from the DNS name.
- **GOOGLE_APPLICATION_CREDENTIALS** is the path to the GCP service account key file.

### Required permissions

The pipeline accesses GCP using a service account configured with the **GOOGLE_APPLICATION_CREDENTIALS** environment variable.
This service account must have GCP permissions equivalent to the following GCP roles:

- Compute Admin (`roles/compute.admin`)
- Kubernetes Engine Admin (`roles/container.admin`)
- Kubernetes Engine Cluster Admin (`roles/container.clusterAdmin`)
- DNS Administrator (`roles/dns.admin`)
- Service Account User (`roles/iam.serviceAccountUser`)
- Storage Admin (`roles/storage.admin`)
