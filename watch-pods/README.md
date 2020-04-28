# Pod State Watcher

## Overview

This tool inspects the state of all Pods.

If at least one Pod is not in the running state or restarts constantly, the program exits with the code other than `0`.
The program runs outside of the cluster so it requires the `kubeconfig` file to access the cluster.

## Prerequisites

To set up the project, download these tools:
* [Go](https://golang.org/dl/) 1.14+
* [Docker](https://www.docker.com/)

These Go and Dep versions are compliant with the `buildpack` used by Prow. For more details read [this](https://github.com/kyma-project/test-infra/blob/master/prow/images/buildpack-golang/README.md) document.

## Usage

This section describes how to build a project image. It also provides the parameters to start and configure a container from the previously built image.

### Build a project image

Execute the `make build build-image` command to perform tests and static code analysis, and to create the `watch-pods` Docker image.

### Configuration options

The following flags are accessible to start and configure the previously built binary file:
- **kubeconfig** (default: in cluster) - a path to the kubeconfig file
- **minWaitingPeriod** (default: 1 minute) - time needed for the cluster to stabilize, after which tests are performed
- **maxWaitingPeriod** (default: 3 minutes) - the maximum period of time in which test are performed
- **reqStabilityPeriod** (default: 1 minute) - the required stability period which informs for how long the container cannot be restarted to be considered as stable

### Run the image
```
docker run --rm --env "ARGS=-maxWaitingPeriod=10m -ignorePodsPattern='core-azure-broker-docs-*'" watch-pods:latest
```
