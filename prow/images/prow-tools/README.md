# Prow Tools

## Overview

The directory contains Dockerfile for prow tools image with prebuilt tools used in prow pipelines.
The image is used as a standalone image in several cleaners job and as additional dependency in `kyma-integration` images.

The image consists of:

- prebuilt binaries from `development/tools/cmd` directory

## Installation
To build the Docker image, run this command:

```bash
make build-image
```