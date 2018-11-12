# Bootstrap Docker Image

## Overview

This project includes the Bootstrap image for Prow infrastructure. Use it for a root image for other Prow images and for generic builds.

This image contains:

- gcloud
- Docker

## Installation

To build the Docker image, run this command:

```bash
docker build bootstrap .
```
