# Cluster

## Overview

The folder contains scripts involved in integration tests.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── cluster-integration                # Scripts for executing the multi-step integration job on a Google Kubernetes Engine (GKE) cluster. This job provisions a cluster, sets up a DNS record, an IP address, and a TLS certificate. It also installs Kyma on the cluster and performs tests on it.
  ├── kind                               # Resources and configuration for the kind cluster
  ├── lib                                # Helper bash scripts for creating pipelines
  ├── build.sh                           # This script builds and tests a given Kyma component by running the respective "Makefile" target.
  ├── governance.sh                      # This script runs the "milv" bot for validating internal and external links in Markdown files. It checks all internal and external links in ".md" files changed in PRs. It also checks all links on the master branch once a day.
  ├── kind-install-kyma.sh               # This script installs and tests Kyma on kind
  ├── kind-upgrade-kyma.sh               # This script tests Kyma upgradeability on kind
  ├── library.sh                         # This script is used as an integral part of other scripts, such as the "build.sh" script. With proper parameters defined, it authenticates you to GCP and sets up the Docker-in-Docker environment.
  ├── provision-vm-and-start-kyma.sh     # This script starts a virtual machine as part of the integration job and runs integration tests for Kyma on Minikube.
  └── publish-buildpack.sh               # This script builds and pushes Docker images for test infrastructure by running the respective "Makefile" target.

```
