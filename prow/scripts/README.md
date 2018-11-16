# Cluster

## Overview

The folder contains scripts involved in integration tests.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── cluster-integration                # Scripts for executing multi-step integration job on GKE cluster. This job provisions a cluster, sets up DNS/IP/Certificates, then installs Kyma and performs tests on the cluster.
  ├── library.sh                         # This script is used as an integral part of other scripts, for example by the "pipeline.sh" script. With proper parameters defined, it authenticates you to GCP and sets up the Docker-in-Docker environment.
  ├── pipeline.sh                        # This script builds and tests a given Kyma component by running the respective "Makefile" target.
  └── provision-vm-and-start-kyma.sh     # This script starts a virtual machine as part of the integration job and runs integration tests for Kyma on Minikube.

```
