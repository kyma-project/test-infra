# Cleaner Docker Image

## Overview
This image contains script which perform cleanup of service account profile in the `kyma-project` project. 
Script requires environment variable `GOOGLE_APPLICATION_CREDENTIALS` which is a path to service account key.

## Build

To build Docker image, run this command:

```bash
docker build -t cleaner:<version> .
```


