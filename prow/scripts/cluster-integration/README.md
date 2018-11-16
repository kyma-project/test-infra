# Cluster Integration

## Overview

The folder contains scripts that are involved in the preparation and removal of the cluster setup on Google Kubernetes Engine (GKE).

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── create-dns-entry.sh               # This script creates a DNS entry on Google Cloud Platform (GCP) that points to the GKE cluster.
  ├── delete-dns-record.sh              # This script removes a DNS record from GCP after completion of integration tests.
  ├── generate-self-signed-cert.sh      # This script creates a self-signed certificate.
  ├── pr-gke-integration.sh             # This script provisions a cluster on GKE and deprovisions it when the tests complete.
  ├── release-ip-address.sh             # This script releases the static IP address of a cluster on GCP.
  └── reserve-ip-address.sh             # This script reserves the static IP address for a cluster on GCP.
```
