# Kubectl Docker Image

## Overview

This folder contains a minimal Docker image used to control the cluster.

This image consists of:

- alpine linux 3.11
- openssl
- curl
- base64
- kubectl (1.17)
- helm (2.16.1)
- helm3 (3.2.1)
- grep

## Installation

To build the Docker image, run this command:

```bash
docker build -t alpine-kubectl .
```
