# Kyma integration images

## Overview

This folder contains the image with tools that are necessary to provision Kyma integration clusters.

The image consists of:
- Go
- Dep
- Helm
- Gcloud
- Az
- Docker

## Maintaining the image

Google Cloud SDK comes with one default `kubectl` version and a couple of additional ones.
See the details [here](https://cloud.google.com/sdk/docs/release-notes#27600_2020-01-14). To set the version of `kubectl` other than the default, run:

- `CLUSTER_VERSION` variable matches the `kubectl` version used in the image with the cluster version that is built by the ProwJob pipeline.

```
ENV CLUSTER_VERSION={VERSION}
RUN mv /google-cloud-sdk/bin/kubectl.${CLUSTER_VERSION} /google-cloud-sdk/bin/kubectl
```

>**CAUTION:** Each image is tagged with the corresponding `kubectl` version. When updating the image, do not forget to adjust the tag. 
```
docker tag $(IMG_NAME) $(IMG):k8s-1.14
```

## Installation

To build the Docker image, run this command:

```bash
make build-image
```
