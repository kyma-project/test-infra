# kind Jobs

This document describes how to execute kind jobs locally without Prow cluster.

## Prerequisites

- [Git](https://git-scm.com/)
- [Docker](https://www.docker.com/)
- [Debian](https://www.debian.org/) machine

## Steps

### Clone repositories

1. Create directory for repositories:
   
   ```bash
   mkdir -p "$HOME/repositories"
   ```

2. Clone `kyma` repository:

   ```bash
   git clone https://github.com/kyma-project/kyma.git "$HOME/repositories/kyma"
   ```

3. Clone `test-infra` repository:

   ```bash
   git clone https://github.com/kyma-project/test-infra.git "$HOME/repositories/test-infra"
   ```

### Start job

1. Create `artifacts` directory:

   ```bash
   mkdir -p "$HOME/artifacts"
   ```

2. Create `docker-graph` directory:

   ```bash
   mkdir -p "$HOME/docker-graph"
   ```

3. Run Docker container with configuration based on ProwJob definition, for example configuration for `pre-master-kyma-kind-upgrade`:

```bash
sudo docker run \
    --rm \
    --privileged \
    --volume "$HOME/repositories:/home/prow/go/src/github.com/kyma-project" \
    --volume "$HOME/artifacts:/artifacts" \
    --volume "/sys/fs/cgroup:/sys/fs/cgroup" \
    --volume "/lib/modules:/lib/modules" \
    --volume "$HOME/docker-graph:/docker-graph" \
    --env ARTIFACTS="/artifacts" \
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