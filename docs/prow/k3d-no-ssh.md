# Run K3d Cluster Inside ProwJobs

This document provides simple instructions, with examples, on how to prepare a ProwJob to use a K3d cluster and Docker.

> This guide also applies to [kind](https://kind.sigs.k8s.io) and other workloads that require Docker.

ProwJob workload operates on a Pod, Kubernetes primitive resource that orchestrates OCI or Docker images. That means
every ProwJob can be configured the same way as any Pod. K3d is a CLI tool that runs a K3s cluster inside the Docker daemon.
Since it's possible to run the Docker daemon inside a Kubernetes Pod (Docker in Docker), this can also be used to run a K3s cluster in it.

Fortunately, in Prow configuration, there are 2 useful presets that configure a ProwJob to use Docker inside. Those 2 presets can be found in this [`config.yaml`](../../prow/config.yaml).

```yaml
labels:
  preset-kind-volume-mounts: "true"
  preset-dind-enabled: "true"
```

Additionally, the ProwJob's container needs to be run in the `privileged` mode since it requires mounting host system paths `/sys/fs/cgroup` and `/lib/modules` to be able to spawn the Docker daemon.

```yaml
securityContext:
  privileged: true
```

Once the ProwJob contains those required definition configurations, you have to explicitly start the Docker service, for example, with `service docker start`, and run the docker-based workloads.

> Remember to use an image with Docker already installed. Otherwise, starting the Docker service will not be possible. For more information on which image can be used, see [prow/images](../../prow/images) directory.

## ProwJob Configuration Example

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
            seccompProfile:
              type: Unconfined
            allowPrivilegeEscalation: true
          resources: # adjust it to your needs
            requests:
              cpu: 500m
              memory: 2Gi
            limits:
              cpu: 1
              memory: 4Gi
```
And the script `ci/test.sh` that is placed in `org/repo`. This script is contained in the target repository, which the job runs against.
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
            seccompProfile:
              type: Unconfined
            allowPrivilegeEscalation: true
          resources: # adjust it to your needs
            requests:
              cpu: 500m
              memory: 2Gi
            limits:
              cpu: 1
              memory: 4Gi
```

## Managing Container Resources

Kubernetes best practices suggest to explicitly define container resource requests and limits.
To ensure that the workload cluster always has the required resources to run the ProwJob, always remember to define `requests` and `limits`
under container `resources` object. We strongly suggest starting with a smaller number, like `cpu: 500m` and `memory: 1Gi`, then adding more if the job seems to be unstable.

## Accessing the Cluster

It's not possible to access the cluster inside the Pod. It's a completely isolated environment without any inbound access from the outside network.
Test all changes locally with Docker.
