# Cluster Integration

## Overview

The folder contains scripts that are responsible for performing a cleanup of Google Cloud Platform (GCP) objects built as part of the integration tests that run on pull requests.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── delete-dns-record.sh            # The script responsible for removing DNS record of type A from a GKE cluster after completion of integration tests.
  ├── pr-gke-integration.sh           # The script that provisions a cluster on GKE and deprovisions it the tests complete.
  └── release-ip-address.sh           # The script which releases the static IP address of a cluster from GCP.
```
