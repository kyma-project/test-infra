# Cleaner Docker Image

## Overview
This image contains the script which performs a cleanup of the service account profile in the `kyma-project` project. 
The script requires the following environment variables:
- **GOOGLE_APPLICATION_CREDENTIALS** which is a path to the service account key.
- **CLOUDSDK_CORE_PROJECT** which is a Gcloud project name.

## Installation

To build the Docker image, run this command:

```bash
docker build -t cleaner:{version} .
```

changed

