# Buildpack Golang with Kubebuilder Docker Image

## Overview

This folder contains the Buildpack Golang with Kubebuilder image that is based on the Golang image. Use it to build components created with Kubebuilder version 2.

The image consists of:

- Kubebuilder 2.0.0
- Kustomize 3.1.0
- Go 1.12
- Dep 0.5.0
- Gcloud 219.0.1
- Docker 18.06.1*

## Installation

To build the Docker image, run this command:

```bash
make build-image
```
