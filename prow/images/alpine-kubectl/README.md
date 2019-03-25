# Kubectl Docker Image

## Overview

This folder contains a minimal docker image that might be used for cluster control. 

This image consists of:

- alpine linux 3.8
- openssl
- curl
- base64
- kubectl (latest)
- helm (2.10)

## Installation

To build the Docker image, run this command:

```bash
docker build alpine-kubectl .
```
