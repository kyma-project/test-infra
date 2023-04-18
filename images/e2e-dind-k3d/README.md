# E2E DinD K3d

This image contains common tools for all jobs/tasks that test Kyma modules in K3d.
K3d runs in docker-in-docker mode. Image is alpine-based.

Usage of this image is restricted only for running K3d/kind workloads for testing.

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

1. Use `securityContext: privileged: true` for this image to be running correctly.
2. Do not define `command` directive in ProwJob or Pod definition. If you have to do it, use `/init.sh` as a value.
3. Pay attention to scripts that run inside the container.
4. To enable Docker-in-Docker you need to use the following presets:
    ```yaml
    preset-dind-enabled: true
    preset-kind-volume-mounts: true
    ```
5. Docker daemon logs will be stored in the `/var/log/dockerd.log` or in `${ARTIFACTS}/dockerd.log` when running from Prow.