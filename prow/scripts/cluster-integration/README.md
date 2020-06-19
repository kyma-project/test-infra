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

> **NOTE:** Run [Job Guard](./../../../development/tools/cmd/jobguard/README.md) at the beginning of each integration job that builds and installs Kyma with changes from the pull request.
  ```bash
  if [[ "${BUILD_TYPE}" == "pr" ]]; then
      shout "Execute Job Guard"
      "${TEST_INFRA_SOURCES_DIR}development/jobguard/scripts/run.sh"
  fi
  ```   

### Project structure

The structure of the folder looks as follows:

```
├── helpers # This directory contains helper scripts used by pipeline jobs.
├── kyma-gke-integration.sh # This script installs and tests Kyma on a real GKE cluster.
├── kyma-gke-nightly.sh # This script creates a long-lived GKE cluster from the master branch. This cluster should be recreated once a day.
├── kyma-gke-upgrade.sh # This script installs the last Kyma release on a GKE cluster and upgrades it with current changes from the PR, master, or release branch. It also triggers the Kyma testing script.
└── kyma-gke-end-to-end-test.sh # This script installs and executes Kyma end-to-end tests on a real GKE cluster.
```

The scripts at the root of the `cluster-integration` directory are used for Prow pipelines. The pipeline uses a toolset from the `Bootstrap` image defined in this repository.

### Required environment variables

This script takes its input configuration from environment variables.
The following environment variables are required:

- **REPO_OWNER** is the repository owner or organization. This variable is set up by Prow.
- **REPO_NAME** is the repository name. This variable is set up by Prow.
- **BUILD_TYPE** is one of pr/master/release. This variable is created by using the `preset-build` label in the Prow job definition.
- **DOCKER_PUSH_REPOSITORY** is the Docker repository hostname.
- **DOCKER_PUSH_DIRECTORY** - the Docker top-level directory, preceded by a slash (/).
- **KYMA_PROJECT_DIR** is a directory path with Kyma sources to use for the installation.
- **CLOUDSDK_CORE_PROJECT** is a Google Cloud Platform (GCP) project for all GCP resources used in the script. For example, the resources include service account, an IP address, a DNS Zone, and an image registry.
- **CLOUDSDK_COMPUTE_REGION** is a GCP compute region.
- **CLOUDSDK_DNS_ZONE_NAME** is a GCP zone name which is different from the DNS name.
- **GOOGLE_APPLICATION_CREDENTIALS** is the path to the GCP service account key file.
- **KYMA_BACKUP_CREDENTIALS** is a secret containing a JSON file with GCP service account credentials. When you configure Velero, use the credentials to grant the Velero server the write and read permissions for the GCP bucket used for backups.
- **KYMA_BACKUP_RESTORE_BUCKET** is a bucket in GCP used to store Kyma's backups.

### Required permissions

The pipeline accesses GCP using a service account configured with the **GOOGLE_APPLICATION_CREDENTIALS** environment variable.
This service account must have GCP permissions equivalent to the following GCP roles:

- Compute Admin (`roles/compute.admin`)
- Kubernetes Engine Admin (`roles/container.admin`)
- Kubernetes Engine Cluster Admin (`roles/container.clusterAdmin`)
- DNS Administrator (`roles/dns.admin`)
- Service Account User (`roles/iam.serviceAccountUser`)
- Storage Admin (`roles/storage.admin`)

### Stackdriver Monitoring

Long-running clusters on GKE use Stackdriver Monitoring to expose some performance metrics collected by Kyma Prometheus instance.
To send metrics to Stackdriver, the collector sidecar container is injected into Prometheus Pod.
Add the container's environment variable to instruct the script to provision cluster with the Stackdriver collector.

    - name: STACKDRIVER_COLLECTOR_SIDECAR_IMAGE_TAG
      value: "0.6.4"
