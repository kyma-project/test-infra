# Kyma Cluster Infra Docker image

## Overview

This folder contains the image with tools useful for managing Kyma clusters, for example for provisioning.

The image consists of:
- go
- dep
- helm

## Installation

To build the Docker image, run this command:

```bash
docker build -t kyma-cluster-infra .
```
