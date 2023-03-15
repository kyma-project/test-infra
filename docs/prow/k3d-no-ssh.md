# Run K3d cluster inside ProwJobs

This document provides simple instruction with examples on how to prepare ProwJob to use K3d cluster and Docker.

> This guide also applies for [kind](https://kind.sigs.k8s.io) and other workloads that require Docker.

ProwJob workload operates on a Pod, Kubernetes' primitive resource that orchestrates OCI or Docker images. That means
every ProwJob can be configured the same way as any Pod. K3d is a CLI tool that runs K3s cluster inside Docker daemon.
Since it's possible to run Docker daemon inside Kubernetes' Pod (Docker in Docker) this can also be used to run K3s cluster in it.

Fortunately, in Prow configuration there are 2 useful presets that configure ProwJob to use Docker inside. Those 2 presets can be found in [prow/config.yaml](../../prow/config.yaml).

```yaml
labels:
  preset-kind-volume-mounts: "true"
  preset-dind-enabled: "true"
```

Additionally, ProwJob's container needs to be run in `privileged` mode, since it requires to mount host's system paths `/sys/fs/cgroup` and `/lib/modules` to be able to spawn Docker daemon.

```yaml
securityContext:
  privileged: true
```

Once ProwJob contains those required definition configurations, you have to explicitly start Docker service, for example with `service docker start` and run their docker-based workloads.

> Remember to use an image with Docker already installed, otherwise starting Docker service will not be possible. For more information which image can be used, see [prow/images](../../prow/images) directory.

## ProwJob configuration example

Here's an example of Docker-powered ProwJob.

```yaml
periodics:
  - name: ci-job-dind
    interval: 1h
    decorate: true
    cluster: untrusted-workload
    labels:
      preset-kind-volume-mounts: "true"
      preset-dind-enabled: "true"
    extra_refs:
      - org: org
        repo: repo
        base_ref: branch
    spec:
      containers:
        - image: eu.gcr.io/kyma-project/test-infra/kyma-integration:v20230119-993f0759
          command:
            - ci/test.sh
          securityContext:
            privileged: true
          resources: # adjust it to your needs
            requests:
              cpu: 500m
              memory: 2Gi
            limits:
              cpu: 1
              memory: 4Gi
```
And the script `ci/test.sh` that is placed in `org/repo`. This script is contained in the target repository, which the job run against.
```sh
service docker start
k3d cluster create
kubectl cluster-info
make test # or any other custom logic
k3d cluster delete 
```

Additionally, you can define the embedded script inside the ProwJob. This approach is good for periodic jobs, where there is no need to run complicated custom scripts.

```yaml
periodics:
  - name: ci-job-dind
    interval: 1h
    decorate: true
    cluster: untrusted-workload
    labels:
      preset-kind-volume-mounts: "true"
      preset-dind-enabled: "true"
    extra_refs:
      - org: org
        repo: repo
        base_ref: branch
    spec:
      containers:
        - image: eu.gcr.io/kyma-project/test-infra/kyma-integration:v20230119-993f0759
          command:
            - /bin/bash
          args:
            - -c
            - |
              service docker start
              k3d cluster create
              kubectl cluster-info
              make test # or any other custom logic
              k3d cluster delete
          securityContext:
            privileged: true
          resources: # adjust it to your needs
            requests:
              cpu: 500m
              memory: 2Gi
            limits:
              cpu: 1
              memory: 4Gi
```

## Managing container resources

Kubernetes best practices suggest to explicitly define container resource requests and limits.
To ensure that workload cluster always has required resources to run the ProwJob always remember to define `requests` and `limits`
under container `resources` object. We strongly suggest to start with smaller number, like `cpu: 500m` and `memory: 1Gi`, then add more if the job seems to be unstable.

## Accessing the cluster

It's not possible to access cluster inside the Pod. It's completely isolated environment without any inbound access from the outside network.
Test all changes locally with Docker.
