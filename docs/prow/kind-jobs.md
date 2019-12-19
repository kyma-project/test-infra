# kind Jobs

This document describes how to run kind jobs locally without a Prow cluster.

## Prerequisites

- [Git](https://git-scm.com/)
- [Docker](https://www.docker.com/)
- [Debian](https://www.debian.org/) machine

## Steps

### Clone repositories

1. Create a directory for the repositories you are about to clone:
   
   ```bash
   mkdir -p "$HOME/repositories"
   ```

2. Clone the `kyma` repository:

   ```bash
   git clone https://github.com/kyma-project/kyma.git "$HOME/repositories/kyma"
   ```

3. Clone the `test-infra` repository:

   ```bash
   git clone https://github.com/kyma-project/test-infra.git "$HOME/repositories/test-infra"
   ```

### Run a kind job

1. Create the `artifacts` directory:

   ```bash
   mkdir -p "$HOME/artifacts"
   ```

2. Create the `docker-graph` directory:

   ```bash
   mkdir -p "$HOME/docker-graph"
   ```

3. Run a Docker container with configuration based on a ProwJob definition. For example, use the configuration for `pre-master-kyma-kind-upgrade`:

```bash
docker run \
    --rm \
    --privileged \
    --volume "$HOME/repositories:/home/prow/go/src/github.com/kyma-project" \
    --volume "$HOME/artifacts:/artifacts" \
    --volume "/sys/fs/cgroup:/sys/fs/cgroup" \
    --volume "/lib/modules:/lib/modules" \
    --volume "$HOME/docker-graph:/docker-graph" \
    --env ARTIFACTS="/artifacts" \
    --env GOPATH="/home/prow/go" \
    eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20191011-51ed45a \
    /home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/kind-upgrade-kyma.sh \
    --ensure-kubectl \
    --ensure-helm \
    --update-hosts \
    --delete-cluster \
    --tune-inotify \
    --start-docker \
    --kubernetes-version \
    v1.14.6 \
    --kyma-sources \
    /home/prow/go/src/github.com/kyma-project/kyma \
    --kyma-installation-timeout \
    30m
```
