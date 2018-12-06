# Cluster

## Overview

The folder contains scripts involved in integration tests.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── cluster-integration                # Scripts for executing the multi-step integration job on a Google Kubernetes Engine (GKE) cluster. This job provisions a cluster, sets up a DNS record, an IP address, and a TLS certificate. It also installs Kyma on the cluster and performs tests on it.
  ├── build.sh                           # This script builds and tests a given Kyma component by running the respective "Makefile" target.
  ├── governance.sh                      # This script runs milv bot for validation internal and external links in markdown files. On PRs checks all internals and externals links on changed markdown files in PR. Also checks all links once a day on master.
  ├── library.sh                         # This script is used as an integral part of other scripts, for example by the "build.sh" script. With proper parameters defined, it authenticates you to GCP and sets up the Docker-in-Docker environment.
  ├── provision-vm-and-start-kyma.sh     # This script starts a virtual machine as part of the integration job and runs integration tests for Kyma on Minikube.
  └── publish-buildpack.sh               # This script builds and pushes Docker images for test infrastructure by running the respective "Makefile" target.

```
