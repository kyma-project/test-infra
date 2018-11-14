# Cluster Integration

## Overview

The folder contains scripts that are involved in the preparation and removal of the cluster setup on Google Kubernetes Engine (GKE). 

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── create-dns-entry.sh               # The script that creates a DNS entry on Google Cloud Platform (GCP) that points to the GKE cluster.
  ├── delete-dns-record.sh              # The script responsible for removing a DNS record from GCP after completion of integration tests.
  ├── generate-self-signed-cert.sh      # The script for creating a self-signed certificate.
  ├── pr-gke-integration.sh             # The script that provisions a cluster on GKE and deprovisions it the tests complete.
  └── release-ip-address.sh             # The script which releases the static IP address of a cluster on GCP.
  └── reserve-ip-address.sh             # The script that reserves the static IP address for a cluster on GCP.
```
