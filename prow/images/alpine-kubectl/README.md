# Kubectl Docker Image

## Overview

This folder contains a minimal Docker image used to control the cluster.

This image consists of:

- alpine linux 3.8
- openssl
- curl
- base64
- kubectl (1.13)
- helm (2.10)

## Installation

To build the Docker image, run this command:

```bash
docker build -t alpine-kubectl .
```
