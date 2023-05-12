# buildpack-go

This image contains a common toolchain for building and running Go-based applications and container images.
It's based on the official Alpine-based [`golang`](https://hub.docker.com/_/golang) image.

This image is suitable for Prow environment.

It contains the following binaries:
* go
* ko
* bash
* curl
* git
* jq
* yq
* wget
* gettext
* ca-certificates
* kustomize
* kubebuilder
* build-base
* dumb-init