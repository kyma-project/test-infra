# Prow Tools

## Overview

The directory contains the Dockerfile for the prow tools image with prebuilt tools used in the prow pipelines.
The image is used as a standalone image in several cleaners jobs and as an additional dependency in the `kyma-integration` images.

The image consists of:

- prebuilt binaries from the `development/tools/cmd` directory

## Installation

To build the Docker image, run this command:

```bash
make build-image
```
