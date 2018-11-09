# Cleaner Docker Image

## Overview
This image contains the script which performs a cleanup of the service account profile in the `kyma-project` project. 
Script requires environment variable `GOOGLE_APPLICATION_CREDENTIALS` which is a path to service account key.

## Installation

To build Docker image, run this command:

```bash
docker build -t cleaner:<version> .
```


