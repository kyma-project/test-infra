# Cluster

## Overview

The folder contains scripts involved in integration tests.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── cluster-integration                # Scripts for the cleanup of Google Cloud Platform (GCP) objects after completion of integration tests.             
  ├── library.sh                         # The script to be used as an integral part of other scripts, for example by the "pipeline.sh" script. With proper parameters defined, it authenticates you to GCP and sets up the Docker-in-Docker environment.
  ├── pipeline.sh                        # The script which builds and tests a given Kyma component by running the respective "Makefile" target.
  └──  provision-vm-and-start-kyma.sh    # The script which starts a virtual machine as part of the integration job and runs integration tests for Kyma on Minikube.

```
