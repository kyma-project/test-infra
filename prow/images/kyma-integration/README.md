# Kyma-integration images

## Overview

This folder contains the image with tools necessary for provisioning kyma-integrtion clusters.

The image consists of:
- Go
- Dep
- Helm
- Gcloud
- Az
- Docker

# Maintaning the image

Google Cloud SDK comes with one default `kubectl` version and couple additional ones.
See details [here](https://cloud.google.com/sdk/docs/release-notes#27600_2020-01-14).

`CLUSTER_VERSION` variable is used to match `kubectl` version used in the image with the cluster version that is build by ProwJob pipeline.

```
ENV CLUSTER_VERSION=1.14
RUN mv /google-cloud-sdk/bin/kubectl.${CLUSTER_VERSION} /google-cloud-sdk/bin/kubectl
```

**CAUTION** Each image is tagged with corresponding `kubectl` version. While updating the image please do not forget to adjust the tag. 
```
docker tag $(IMG_NAME) $(IMG):k8s-1.14
```

## Installation

To build the Docker image, run this command:

```bash
make build-image
```
