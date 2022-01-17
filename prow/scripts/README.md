# Cluster

## Overview

The folder contains scripts involved in integration tests.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── cluster-integration                         # Scripts for executing the multi-step integration job on a Google Kubernetes Engine (GKE) cluster. This job provisions a cluster, sets up a DNS record, an IP address, and a TLS certificate. It also installs Kyma on the cluster and performs tests on it.
  ├── kind                                        # Resources and configuration for the kind cluster
  ├── lib                                         # Helper bash scripts for creating pipelines
  ├── resources                                   # Files used directly by pipelines
  ├── governance.sh                               # This script runs the "milv" bot for validating internal and external links in Markdown files. It checks all internal and external links in ".md" files changed in PRs. It also checks all links on the main branch once a day.
  ├── provision-vm-and-start-kyma-minikube.sh     # This script starts a virtual machine as part of the integration job and runs integration tests for Kyma on Minikube.
  ├── provision-vm-and-start-kyma-k3d.sh          # This script starts a virtual machine as part of the integration job and runs fast integration tests for Kyma on k3d.
  ├── validate-config.sh                          # This script runs the "Checker" application and checks the uniqueness of jobs names.
  ├── validate-scripts.sh                         # This script performs a static analysis of bash scripts in the "test-infra" repository.
  └── publish-buildpack.sh                        # This script builds and pushes Docker images for test infrastructure by running the respective "Makefile" target.


```
