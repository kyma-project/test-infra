# Bootstrap Docker Image

## Overview

This project includes the Bootstrap image for Prow infrastructure. It can be used as a root image for other Prow images and also for generic builds.

This image contains:

- gcloud
- Docker

## Build

To build Docker image, run this command:

```bash
docker build bootstrap .
```
