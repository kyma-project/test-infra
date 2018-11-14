# Cluster

## Overview

The folder contains various scripts involved in integration tests.

### Project structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── cluster-integration                # Scripts for the cleanup of Google Cloud Platform (GCP) objects after completion of integration tests.             
  ├── library.sh                         # The script to be used as an integral part of other scripts. For example, the "pipeline.sh" script uses it. With proper parameters defined, it authenticates you in GCP and sets up the Docker-in-Docker environment.
  ├── pipeline.sh                        # The script responsible for building and testing a given Kyma component by running the respective "Makefile" target.
  └──  provision-vm-and-start-kyma.sh    # The script responsible for starting a virtual machine as part of the integration job and running integration tests for Kyma on Minikube.

```
