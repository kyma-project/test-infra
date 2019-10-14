# kind Environment

## Overview

[`kind`](https://github.com/kubernetes-sigs/kind) is a tool for running local Kubernetes clusters with nodes deployed as Docker containers. It can be deployed as a multi-node container, which gives us an ability to test Kyma components in an environment similar to the production cluster.

### Directory structure

The `kind` folder contains a set of configuration files for the cluster built with `kind`.

Its structure looks as follows:

```
  ├── cluster                  # Cluster configuration
        └── resources          # Default resources applied during cluster setup
  ├── config                   # Kyma overrides for "kind"
  ├── scripts                  # Scripts used for the "kind" environment
  └── KIND_KUBERNETES_VERSION  # File with a particular Kubernetes version used for the cluster
```

### Deploy a Kyma cluster

To deploy Kyma on `kind`, run:

```bash
./scripts/install-kyma.sh
```

To list available options, run:

```bash
./scripts/install-kyma.sh --help
```

It may require more resources for Docker, adjust it in Docker preferences (cpu and memory)

### Build a Kubernetes image

Run the `scripts/build-kubernetes-image.sh` script to build a Kubernetes image used in `kind`. That script is specifically designed for the Prow pipeline.

To create your own Kubernetes image, follow [this](https://kind.sigs.k8s.io/docs/user/quick-start/#building-images) instruction.
