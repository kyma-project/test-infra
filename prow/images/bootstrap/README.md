# Bootstrap Docker Image

## Overview

This folder contains the Bootstrap image for Prow infrastructure. Use it for a root image for other Prow images and for generic builds.

The image consists of:

- curl
- dnsutils
- Docker
- gcloud
- gettext
- git
- jq
- python
- shellcheck
- wget

## Installation

To build the Docker image, run this command:

```bash
docker build bootstrap .
```
