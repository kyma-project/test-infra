# E2E DinD K3d Image

This image contains common tools for all jobs/tasks that test Kyma modules in K3d.
K3d runs in the Docker-in-Docker mode. The image is Alpine-based.

The usage of this image is restricted only to running K3d/kind workloads for testing.

It contains the following binaries:

* go
* k3d
* kind
* helm
* build-base
* ca-certificates
* curl
* bash
* jq
* yq
* docker
* docker-compose
* btrfs-progs
* e2fsprogs
* e2fsprogs-extra
* iptables
* xfsprogs
* xz
* git
* fuse-overlayfs
* device-mapper
* openssh-client
* dumb-init

## Usage

1. Use `securityContext: privileged: true` for this image to run correctly.
2. Do not define the `command` directive in ProwJob or Pod definition. If you have to do it, use `/init.sh` as a value.
3. Pay attention to scripts that run inside the container.
4. To enable Docker-in-Docker you need to use the following presets:
    ```yaml
    preset-dind-enabled: true
    preset-kind-volume-mounts: true
    ```
5. Docker daemon logs are stored in the `/var/log/dockerd.log` or in `${ARTIFACTS}/dockerd.log` when running from Prow.