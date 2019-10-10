# kind Environment

## Overview

kind is a tool for running local Kubernetes cluster with nodes deployed as Docker containers. It can be deployed as a multi-node container what gives as ability to test in a environment close to production cluster.

## Directory structure

The `kind` folder contains a set of configuration files for the kind cluster.

Its structure looks as follows:

```
  ├── cluster                  # kind cluster configuration
        └── resources          # Default resources applied during cluster setup
  ├── config                   # Kyma overrides for kind
  ├── scripts                  # Scripts used for kind environment
  └── KIND_KUBERNETES_VERSION  # File with Kubernetes version used for cluster
```

## Deploy Kyma cluster on kind

To deploy Kyma on kind, execute:

```bash
./scripts/install-kyma.sh
```

To list available options, execute:

```bash
./scripts/install-kyma.sh --help
```

It may require more resources for Docker, adjust it in Docker preferences (cpu and memory)

## Build Kubernetes image

To build a Kubernetes image that is used in Kind, script `scripts/build-kubernetes-image.sh` is used. That script is designed to be used on Prow pipeline.

For creating your own Kubernetes image, follow https://kind.sigs.k8s.io/docs/user/quick-start/#building-images.