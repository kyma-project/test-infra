# Performance Cluster


### Description: Kyma Upgradeability plan on GKE. The purpose of this script is to install last Kyma release on real GKE cluster, upgrade it with current changes and trigger testing.

## Expected vars:

- INPUT_CLUSTER_NAME - name for the new cluster
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

### Permissions: 

In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
- Compute Admin
- Kubernetes Engine Admin
- Kubernetes Engine Cluster Admin
- DNS Administrator
- Service Account User
- Storage Admin
- Compute Network Admin
